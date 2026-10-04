package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type CreateTask struct {
	store      TaskStore
	events     EventLog
	transactor Transactor
	projects   ProjectChecker
	spheres    SphereChecker
	rules      *DomainRules
	now        func() time.Time
}

func NewCreateTask(store TaskStore, events EventLog, transactor Transactor, projects ProjectChecker, spheres SphereChecker) *CreateTask {
	return &CreateTask{
		store: store, events: events, transactor: transactor, projects: projects, spheres: spheres,
		now: func() time.Time { return time.Now().UTC() },
	}
}

// WithDomainRules подключает домен-правила (TASK-011 п.8) после создания задачи.
func (uc *CreateTask) WithDomainRules(rules *DomainRules) *CreateTask {
	uc.rules = rules
	return uc
}

type CreateTaskInput struct {
	UserID          ids.UserID
	Title           string
	Description     *string
	Priority        domain.Priority
	Kind            domain.Kind
	Address         *string
	NoteID          *ids.NoteID
	DueDate         *time.Time
	DurationMinutes *int
	Tags            []string
	ProjectIDs      []ids.ProjectID
	SphereIDs       []ids.SphereID
	Source          events.Source
}

func (uc *CreateTask) Execute(ctx context.Context, in CreateTaskInput) (TaskDTO, error) {
	if in.UserID.IsZero() {
		return TaskDTO{}, fmt.Errorf("user id is required")
	}
	if in.Priority == "" {
		in.Priority = domain.PriorityMedium
	}

	title := in.Title
	tags := domain.NormalizeTags(in.Tags)
	if clean, fromTitle := domain.ExtractHashtags(title); len(fromTitle) > 0 {
		title = clean
		tags = domain.NormalizeTags(append(tags, fromTitle...))
	}

	now := uc.now()
	task, err := domain.NewTask(in.UserID, title, in.Priority, in.DueDate, now)
	if err != nil {
		return TaskDTO{}, err
	}
	if in.Description != nil {
		desc := strings.TrimSpace(*in.Description)
		if desc != "" {
			task.Description = &desc
		}
	}
	if in.DurationMinutes != nil {
		if *in.DurationMinutes <= 0 {
			return TaskDTO{}, domain.ErrInvalidDuration
		}
		mins := *in.DurationMinutes
		task.DurationMinutes = &mins
	}
	task.Tags = tags
	if in.Kind != "" {
		if !in.Kind.Valid() {
			return TaskDTO{}, domain.ErrInvalidKind
		}
		task.Kind = in.Kind
	}
	if in.Address != nil {
		addr := strings.TrimSpace(*in.Address)
		if addr != "" {
			task.Address = &addr
		}
	}
	if in.NoteID != nil && !in.NoteID.IsZero() {
		id := *in.NoteID
		task.NoteID = &id
	}
	if len(in.ProjectIDs) > 0 && uc.projects != nil {
		ok, err := uc.projects.AllExist(ctx, in.UserID, in.ProjectIDs)
		if err != nil {
			return TaskDTO{}, fmt.Errorf("validate projects: %w", err)
		}
		if !ok {
			return TaskDTO{}, fmt.Errorf("project not found")
		}
		task.ProjectIDs = in.ProjectIDs
	}
	if len(in.SphereIDs) > 0 && uc.spheres != nil {
		ok, err := uc.spheres.AllExist(ctx, in.UserID, in.SphereIDs)
		if err != nil {
			return TaskDTO{}, fmt.Errorf("validate spheres: %w", err)
		}
		if !ok {
			return TaskDTO{}, fmt.Errorf("sphere not found")
		}
		task.SphereIDs = in.SphereIDs
	}

	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.Save(txCtx, task); err != nil {
			return err
		}
		if len(task.ProjectIDs) > 0 {
			if err := uc.store.SetProjects(txCtx, task.ID, task.ProjectIDs); err != nil {
				return err
			}
		}
		if len(task.SphereIDs) > 0 {
			if err := uc.store.SetSpheres(txCtx, task.ID, task.SphereIDs); err != nil {
				return err
			}
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        task.UserID,
			AggregateType: "task",
			AggregateID:   task.ID.UUID(),
			EventType:     "TaskCreated",
			Payload: map[string]any{
				"title": task.Title, "description": task.Description,
				"priority": task.Priority, "due_date": task.DueDate,
				"duration_minutes": task.DurationMinutes, "tags": task.Tags, "project_ids": task.ProjectIDs, "sphere_ids": task.SphereIDs,
			},
			Source:     in.Source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return TaskDTO{}, fmt.Errorf("create task: %w", err)
	}

	dto := ToDTO(task)
	if uc.rules != nil {
		// best-effort: правило не должно отменять уже созданную задачу
		_ = uc.rules.Apply(ctx, in.UserID, dto)
	}
	return dto, nil
}
