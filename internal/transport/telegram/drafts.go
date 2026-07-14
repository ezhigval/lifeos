package telegram

import (
	"fmt"
	"html"
	"strings"

	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

const (
	CBSettingsAdd      = "settings:add"
	CBSettingsDel      = "settings:del:"
	CBDraftSphere      = "draft:sphere:"
	CBDraftSphereOK    = "draft:sphere:ok"
	CBDraftProject     = "draft:project:"
	CBDraftProjectOK   = "draft:project:ok"
	CBDraftProjectSkip = "draft:project:skip"

	PayloadDraftProjectName    = "draft_project_name"
	PayloadDraftProjectSpheres = "draft_project_spheres"
	PayloadDraftTaskTitle      = "draft_task_title"
	PayloadDraftTaskProjects   = "draft_task_projects"
	PayloadCurrentSection      = "current_section"
)

func FormatSectionHeader(section, body string) string {
	if body == "" {
		return fmt.Sprintf("<b>%s</b>", html.EscapeString(section))
	}
	return fmt.Sprintf("<b>%s</b>\n%s", html.EscapeString(section), body)
}

func InlineSettingsActions() [][]InlineButton {
	return [][]InlineButton{{{Text: "➕ Добавить сферу", CallbackData: CBSettingsAdd}}}
}

func FormatSettingsSpheres(items []spheresapp.SphereDTO) (string, [][]InlineButton) {
	text := FormatSpheres(items)
	if len(items) == 0 {
		text += "\n\nДобавь первую сферу кнопкой ниже."
		return text, InlineSettingsActions()
	}
	var rows [][]InlineButton
	for _, item := range items {
		rows = append(rows, []InlineButton{{
			Text:         "🗑 " + truncate(item.Name, 28),
			CallbackData: CBSettingsDel + item.ID.String(),
		}})
	}
	return text, PrependInline(InlineSettingsActions(), rows)
}

func FormatSpherePicker(name string, spheres []spheresapp.SphereDTO, selected map[string]bool) (string, [][]InlineButton) {
	var b strings.Builder
	fmt.Fprintf(&b, "📁 <b>Новый проект</b>\nНазвание: <b>%s</b>\n\nВыбери сферы (можно несколько):", html.EscapeString(name))
	var rows [][]InlineButton
	for _, s := range spheres {
		mark := "○"
		if selected[s.ID.String()] {
			mark = "✓"
		}
		rows = append(rows, []InlineButton{{
			Text:         mark + " " + truncate(s.Name, 28),
			CallbackData: CBDraftSphere + s.ID.String(),
		}})
	}
	rows = append(rows, []InlineButton{{Text: "✅ Создать проект", CallbackData: CBDraftSphereOK}})
	return strings.TrimSpace(b.String()), rows
}

func FormatProjectPicker(title string, projects []projectsapp.ProjectDTO, selected map[string]bool) (string, [][]InlineButton) {
	var b strings.Builder
	fmt.Fprintf(&b, "✅ <b>Новая задача</b>\nНазвание: <b>%s</b>\n\nВыбери проекты (можно несколько):", html.EscapeString(title))
	var rows [][]InlineButton
	for _, p := range projects {
		mark := "○"
		if selected[p.ID.String()] {
			mark = "✓"
		}
		rows = append(rows, []InlineButton{{
			Text:         mark + " " + truncate(p.Name, 28),
			CallbackData: CBDraftProject + p.ID.String(),
		}})
	}
	rows = append(rows, []InlineButton{
		{Text: "✅ Создать задачу", CallbackData: CBDraftProjectOK},
		{Text: "Без проекта", CallbackData: CBDraftProjectSkip},
	})
	if len(projects) == 0 {
		b.WriteString("\n\nПроектов пока нет — создай задачу без привязки.")
	}
	return strings.TrimSpace(b.String()), rows
}

func idsFromPayload(raw any) map[string]bool {
	out := map[string]bool{}
	arr, ok := raw.([]any)
	if !ok {
		return out
	}
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			out[s] = true
		}
	}
	return out
}

func toggleID(list []string, id string) []string {
	for i, v := range list {
		if v == id {
			return append(list[:i], list[i+1:]...)
		}
	}
	return append(list, id)
}

func parseSphereIDs(list []string) ([]ids.SphereID, error) {
	out := make([]ids.SphereID, 0, len(list))
	for _, s := range list {
		id, err := ids.ParseSphereID(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func parseProjectIDs(list []string) ([]ids.ProjectID, error) {
	out := make([]ids.ProjectID, 0, len(list))
	for _, s := range list {
		id, err := ids.ParseProjectID(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func payloadStringSlice(payload map[string]any, key string) []string {
	raw, ok := payload[key]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func stringSliceToAny(list []string) []any {
	out := make([]any, len(list))
	for i, s := range list {
		out[i] = s
	}
	return out
}
