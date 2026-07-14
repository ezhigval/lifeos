package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type GetTask struct {
	store TaskStore
}

func NewGetTask(store TaskStore) *GetTask {
	return &GetTask{store: store}
}

func (uc *GetTask) Execute(ctx context.Context, userID ids.UserID, taskID ids.TaskID) (TaskDTO, error) {
	if userID.IsZero() || taskID.IsZero() {
		return TaskDTO{}, fmt.Errorf("user id and task id are required")
	}
	task, err := uc.store.GetByID(ctx, userID, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return TaskDTO{}, ErrTaskNotFound
		}
		return TaskDTO{}, err
	}
	return ToDTO(task), nil
}

type UpdateTask struct {
	store      TaskStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewUpdateTask(store TaskStore, events EventLog, transactor Transactor) *UpdateTask {
	return &UpdateTask{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type UpdateTaskInput struct {
	UserID      ids.UserID
	TaskID      ids.TaskID
	Title       string
	Priority    domain.Priority
	DueDate     *time.Time
	Description *string
	Source      events.Source
}

func (uc *UpdateTask) Execute(ctx context.Context, in UpdateTaskInput) (TaskDTO, error) {
	if in.UserID.IsZero() || in.TaskID.IsZero() {
		return TaskDTO{}, fmt.Errorf("user id and task id are required")
	}
	if in.Priority == "" {
		in.Priority = domain.PriorityMedium
	}

	now := uc.now()
	var result TaskDTO
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		task, err := uc.store.GetByID(txCtx, in.UserID, in.TaskID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return ErrTaskNotFound
			}
			return err
		}
		if err := task.ApplyEdit(in.Title, in.Priority, in.DueDate, in.Description, now); err != nil {
			return err
		}
		if err := uc.store.Update(txCtx, task); err != nil {
			return err
		}
		if err := uc.events.Append(txCtx, events.Record{
			UserID:        task.UserID,
			AggregateType: "task",
			AggregateID:   task.ID.UUID(),
			EventType:     "TaskUpdated",
			Payload: map[string]any{
				"title": task.Title, "priority": task.Priority, "due_date": task.DueDate,
			},
			Source:     in.Source,
			OccurredAt: now,
		}); err != nil {
			return err
		}
		result = ToDTO(task)
		return nil
	})
	if err != nil {
		return TaskDTO{}, fmt.Errorf("update task: %w", err)
	}
	return result, nil
}

type ArchiveTask struct {
	store      TaskStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewArchiveTask(store TaskStore, events EventLog, transactor Transactor) *ArchiveTask {
	return &ArchiveTask{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type ArchiveTaskInput struct {
	UserID ids.UserID
	TaskID ids.TaskID
	Source events.Source
}

func (uc *ArchiveTask) Execute(ctx context.Context, in ArchiveTaskInput) (TaskDTO, error) {
	if in.UserID.IsZero() || in.TaskID.IsZero() {
		return TaskDTO{}, fmt.Errorf("user id and task id are required")
	}
	now := uc.now()
	var result TaskDTO
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		task, err := uc.store.GetByID(txCtx, in.UserID, in.TaskID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return ErrTaskNotFound
			}
			return err
		}
		if err := task.Archive(now); err != nil {
			return err
		}
		if err := uc.store.Update(txCtx, task); err != nil {
			return err
		}
		if err := uc.events.Append(txCtx, events.Record{
			UserID:        task.UserID,
			AggregateType: "task",
			AggregateID:   task.ID.UUID(),
			EventType:     "TaskArchived",
			Payload:       map[string]any{"status": task.Status},
			Source:        in.Source,
			OccurredAt:    now,
		}); err != nil {
			return err
		}
		result = ToDTO(task)
		return nil
	})
	if err != nil {
		return TaskDTO{}, fmt.Errorf("archive task: %w", err)
	}
	return result, nil
}

type DeleteTask struct {
	store      TaskStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewDeleteTask(store TaskStore, events EventLog, transactor Transactor) *DeleteTask {
	return &DeleteTask{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type DeleteTaskInput struct {
	UserID ids.UserID
	TaskID ids.TaskID
	Source events.Source
}

func (uc *DeleteTask) Execute(ctx context.Context, in DeleteTaskInput) error {
	if in.UserID.IsZero() || in.TaskID.IsZero() {
		return fmt.Errorf("user id and task id are required")
	}
	now := uc.now()
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		task, err := uc.store.GetByID(txCtx, in.UserID, in.TaskID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return ErrTaskNotFound
			}
			return err
		}
		if err := task.SoftDelete(now); err != nil {
			return err
		}
		if err := uc.store.Update(txCtx, task); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        task.UserID,
			AggregateType: "task",
			AggregateID:   task.ID.UUID(),
			EventType:     "TaskDeleted",
			Payload:       map[string]any{"deleted_at": task.DeletedAt},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}
