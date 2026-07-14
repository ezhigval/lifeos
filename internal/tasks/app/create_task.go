package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type TaskStore interface {
	Save(ctx context.Context, task domain.Task) error
	SetProjects(ctx context.Context, taskID ids.TaskID, projectIDs []ids.ProjectID) error
	GetByID(ctx context.Context, userID ids.UserID, taskID ids.TaskID) (domain.Task, error)
	ListByDueDate(ctx context.Context, userID ids.UserID, dueDate time.Time) ([]domain.Task, error)
	ListByProject(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) ([]domain.Task, error)
	FindOpenByTitle(ctx context.Context, userID ids.UserID, title string) (domain.Task, error)
	Update(ctx context.Context, task domain.Task) error
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type TaskDTO struct {
	ID         ids.TaskID
	Title      string
	Status     domain.Status
	Priority   domain.Priority
	DueDate    *time.Time
	ProjectIDs []ids.ProjectID
}

func ToDTO(task domain.Task) TaskDTO {
	idsCopy := append([]ids.ProjectID(nil), task.ProjectIDs...)
	return TaskDTO{
		ID: task.ID, Title: task.Title, Status: task.Status, Priority: task.Priority,
		DueDate: task.DueDate, ProjectIDs: idsCopy,
	}
}

func ToDTOs(tasks []domain.Task) []TaskDTO {
	out := make([]TaskDTO, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, ToDTO(task))
	}
	return out
}

type ProjectChecker interface {
	AllExist(ctx context.Context, userID ids.UserID, projectIDs []ids.ProjectID) (bool, error)
}

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
	UserID     ids.UserID
	Title      string
	Priority   domain.Priority
	DueDate    *time.Time
	ProjectIDs []ids.ProjectID
	Source     events.Source
}

func (uc *CreateTask) Execute(ctx context.Context, in CreateTaskInput) (TaskDTO, error) {
	if in.UserID.IsZero() {
		return TaskDTO{}, fmt.Errorf("user id is required")
	}
	if in.Priority == "" {
		in.Priority = domain.PriorityMedium
	}

	now := uc.now()
	task, err := domain.NewTask(in.UserID, in.Title, in.Priority, in.DueDate, now)
	if err != nil {
		return TaskDTO{}, err
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
				"title": task.Title, "priority": task.Priority, "due_date": task.DueDate, "project_ids": task.ProjectIDs,
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
