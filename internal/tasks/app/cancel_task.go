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

type CancelTask struct {
	store      TaskStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCancelTask(store TaskStore, events EventLog, transactor Transactor) *CancelTask {
	return &CancelTask{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CancelTaskInput struct {
	UserID ids.UserID
	TaskID ids.TaskID
	Source events.Source
}

func (uc *CancelTask) Execute(ctx context.Context, in CancelTaskInput) (TaskDTO, error) {
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
		if err := task.Cancel(now); err != nil {
			return err
		}
		if err := uc.store.Update(txCtx, task); err != nil {
			return err
		}
		if err := uc.events.Append(txCtx, events.Record{
			UserID:        task.UserID,
			AggregateType: "task",
			AggregateID:   task.ID.UUID(),
			EventType:     "TaskCancelled",
			Payload:       map[string]any{"title": task.Title},
			Source:        in.Source,
			OccurredAt:    now,
		}); err != nil {
			return err
		}
		result = ToDTO(task)
		return nil
	})
	if err != nil {
		return TaskDTO{}, fmt.Errorf("cancel task: %w", err)
	}
	return result, nil
}
