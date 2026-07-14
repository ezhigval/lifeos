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
	now        func() time.Time
}

func NewCreateTask(store TaskStore, events EventLog, transactor Transactor, projects ProjectChecker) *CreateTask {
	return &CreateTask{
		store: store, events: events, transactor: transactor, projects: projects,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateTaskInput struct {
	UserID          ids.UserID
	Title           string
	Description     *string
	Priority        domain.Priority
	DueDate         *time.Time
	DurationMinutes *int
	Tags            []string
	ProjectIDs      []ids.ProjectID
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

	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.Save(txCtx, task); err != nil {
			return err
		}
		if len(task.ProjectIDs) > 0 {
			if err := uc.store.SetProjects(txCtx, task.ID, task.ProjectIDs); err != nil {
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
				"duration_minutes": task.DurationMinutes, "tags": task.Tags, "project_ids": task.ProjectIDs,
			},
			Source:     in.Source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return TaskDTO{}, fmt.Errorf("create task: %w", err)
	}

	return ToDTO(task), nil
}
