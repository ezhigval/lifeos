package telegram

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	spheresdomain "github.com/valentinezhov/lifeos/internal/spheres/domain"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
	tginfra "github.com/valentinezhov/lifeos/internal/transport/telegram/infra"
)

func (h *MessageHandler) basePayload(ctx context.Context, userID ids.UserID) map[string]any {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil || sess.StatePayload == nil {
		return map[string]any{}
	}
	out := map[string]any{}
	if v, ok := sess.StatePayload[replyKBSetKey]; ok {
		out[replyKBSetKey] = v
	}
	if v, ok := sess.StatePayload[replyKBVersionKey]; ok {
		out[replyKBVersionKey] = v
	}
	if v, ok := sess.StatePayload["view_project_id"]; ok {
		out["view_project_id"] = v
	}
	return out
}

func (h *MessageHandler) onProjectTitleEntered(ctx context.Context, userID ids.UserID, title string) (dispatchResult, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return dispatchResult{text: "Название проекта не может быть пустым."}, nil
	}
	payload := h.basePayload(ctx, userID)
	payload[PayloadDraftProjectName] = title
	payload[PayloadDraftProjectSpheres] = []any{}
	if err := h.sessions.SetState(ctx, userID, tginfra.StateAwaitProjectSpheres, payload); err != nil {
		return dispatchResult{}, err
	}
	return h.showSpherePicker(ctx, userID)
}

func (h *MessageHandler) showSpherePicker(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	name, _ := sess.StatePayload[PayloadDraftProjectName].(string)
	spheres, err := h.listSpheres.Execute(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	if len(spheres) == 0 {
		return dispatchResult{text: "Сначала добавь сферу в ⚙️ Настройках."}, nil
	}
	selected := idsFromPayload(sess.StatePayload[PayloadDraftProjectSpheres])
	text, inline := FormatSpherePicker(name, spheres, selected)
	return dispatchResult{text: text, inline: inline}, nil
}

func (h *MessageHandler) toggleDraftSphere(ctx context.Context, userID ids.UserID, sphereID string) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	list := payloadStringSlice(sess.StatePayload, PayloadDraftProjectSpheres)
	list = toggleID(list, sphereID)
	if err := h.sessions.UpdatePayload(ctx, userID, map[string]any{
		PayloadDraftProjectSpheres: stringSliceToAny(list),
	}); err != nil {
		return dispatchResult{}, err
	}
	return h.showSpherePicker(ctx, userID)
}

func (h *MessageHandler) confirmDraftProject(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	name, _ := sess.StatePayload[PayloadDraftProjectName].(string)
	sphereIDs, err := parseSphereIDs(payloadStringSlice(sess.StatePayload, PayloadDraftProjectSpheres))
	if err != nil {
		return dispatchResult{}, err
	}
	if len(sphereIDs) == 0 {
		picker, err := h.showSpherePicker(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		picker.text += "\n\n⚠️ Выбери хотя бы одну сферу."
		return picker, nil
	}
	dto, err := h.createProject.Execute(ctx, projectsapp.CreateProjectInput{
		UserID: userID, Name: name, SphereIDs: sphereIDs, Source: events.SourceTelegram,
	})
	if err != nil {
		return dispatchResult{}, err
	}
	_ = h.sessions.SetState(ctx, userID, tginfra.StateIdle, h.basePayload(ctx, userID))
	projects, err := h.projectsPickerView(ctx, userID)
	if err != nil {
		return dispatchResult{text: FormatProjectCreated(dto)}, nil
	}
	return dispatchResult{
		text:   FormatProjectCreated(dto) + "\n\n" + projects.text,
		inline: projects.inline,
	}, nil
}

func (h *MessageHandler) onTaskTitleEntered(ctx context.Context, userID ids.UserID, title string) (dispatchResult, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return dispatchResult{text: "Название задачи не может быть пустым."}, nil
	}
	payload := h.basePayload(ctx, userID)
	payload[PayloadDraftTaskTitle] = title
	payload[PayloadDraftTaskProjects] = []any{}
	if err := h.sessions.SetState(ctx, userID, tginfra.StateAwaitTaskProjects, payload); err != nil {
		return dispatchResult{}, err
	}
	return h.showProjectPickerForTask(ctx, userID)
}

func (h *MessageHandler) showProjectPickerForTask(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	title, _ := sess.StatePayload[PayloadDraftTaskTitle].(string)
	projects, err := h.listProjects.Execute(ctx, projectsapp.ListProjectsInput{UserID: userID})
	if err != nil {
		return dispatchResult{}, err
	}
	selected := idsFromPayload(sess.StatePayload[PayloadDraftTaskProjects])
	text, inline := FormatProjectPicker(title, projects, selected)
	return dispatchResult{text: text, inline: inline}, nil
}

func (h *MessageHandler) toggleDraftProject(ctx context.Context, userID ids.UserID, projectID string) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	list := payloadStringSlice(sess.StatePayload, PayloadDraftTaskProjects)
	list = toggleID(list, projectID)
	if err := h.sessions.UpdatePayload(ctx, userID, map[string]any{
		PayloadDraftTaskProjects: stringSliceToAny(list),
	}); err != nil {
		return dispatchResult{}, err
	}
	return h.showProjectPickerForTask(ctx, userID)
}

func (h *MessageHandler) confirmDraftTask(ctx context.Context, userID ids.UserID, projectIDs []ids.ProjectID) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	title, _ := sess.StatePayload[PayloadDraftTaskTitle].(string)
	if title == "" {
		return dispatchResult{text: "Черновик задачи потерян. Нажми ➕ Задача снова."}, nil
	}
	tz, err := h.tzReader.Timezone(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	today, err := timeutil.DateInTimezone(time.Now().UTC(), tz)
	if err != nil {
		return dispatchResult{}, err
	}
	dto, err := h.createTask.Execute(ctx, tasksapp.CreateTaskInput{
		UserID: userID, Title: title, Priority: taskdomain.PriorityMedium,
		DueDate: &today, ProjectIDs: projectIDs, Source: events.SourceTelegram,
	})
	if err != nil {
		return dispatchResult{}, err
	}
	_ = h.sessions.SetState(ctx, userID, tginfra.StateIdle, h.basePayload(ctx, userID))
	tasks, err := h.tasksTodayView(ctx, userID)
	if err != nil {
		return dispatchResult{text: FormatTaskCreated(dto)}, nil
	}
	return dispatchResult{
		text:   FormatTaskCreated(dto) + "\n\n" + tasks.text,
		inline: tasks.inline,
	}, nil
}

func (h *MessageHandler) confirmDraftTaskWithSelection(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	projectIDs, err := parseProjectIDs(payloadStringSlice(sess.StatePayload, PayloadDraftTaskProjects))
	if err != nil {
		return dispatchResult{}, err
	}
	return h.confirmDraftTask(ctx, userID, projectIDs)
}

func (h *MessageHandler) confirmDraftTaskSkip(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	return h.confirmDraftTask(ctx, userID, nil)
}

func (h *MessageHandler) beginAddSphere(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	payload := h.basePayload(ctx, userID)
	if err := h.sessions.SetState(ctx, userID, tginfra.StateAwaitSphereName, payload); err != nil {
		return dispatchResult{}, err
	}
	return dispatchResult{text: "Введи название новой сферы:"}, nil
}

func (h *MessageHandler) onSphereNameEntered(ctx context.Context, userID ids.UserID, name string) (dispatchResult, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return dispatchResult{text: "Название сферы не может быть пустым."}, nil
	}
	dto, err := h.createSphere.Execute(ctx, spheresapp.CreateSphereInput{
		UserID: userID, Name: name, Source: events.SourceTelegram,
	})
	if err != nil {
		return dispatchResult{}, err
	}
	_ = h.sessions.SetState(ctx, userID, tginfra.StateIdle, h.basePayload(ctx, userID))
	settings, err := h.settingsView(ctx, userID)
	if err != nil {
		return dispatchResult{text: FormatSphereCreated(dto)}, nil
	}
	return dispatchResult{
		text:   FormatSphereCreated(dto) + "\n\n" + settings.text,
		inline: settings.inline,
	}, nil
}

func (h *MessageHandler) deleteSphereByID(ctx context.Context, userID ids.UserID, rawID string) (dispatchResult, error) {
	sphereID, err := ids.ParseSphereID(rawID)
	if err != nil {
		return dispatchResult{}, err
	}
	dto, err := h.deleteSphere.Execute(ctx, spheresapp.DeleteSphereInput{
		UserID: userID, SphereID: sphereID, Source: events.SourceTelegram,
	})
	if err != nil {
		if errors.Is(err, spheresapp.ErrSphereNotFound) {
			return h.settingsView(ctx, userID)
		}
		if errors.Is(err, spheresdomain.ErrHasProjects) {
			msg := "Нельзя удалить сферу с привязанными проектами."
			settings, serr := h.settingsView(ctx, userID)
			if serr != nil {
				return dispatchResult{text: msg}, nil
			}
			return dispatchResult{text: msg + "\n\n" + settings.text, inline: settings.inline}, nil
		}
		return dispatchResult{}, err
	}
	settings, err := h.settingsView(ctx, userID)
	if err != nil {
		return dispatchResult{text: FormatSphereDeleted(dto)}, nil
	}
	return dispatchResult{
		text:   FormatSphereDeleted(dto) + "\n\n" + settings.text,
		inline: settings.inline,
	}, nil
}
