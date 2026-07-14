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

type RescheduleTask struct {
	store      TaskStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewRescheduleTask(store TaskStore, events EventLog, transactor Transactor) *RescheduleTask {
	return &RescheduleTask{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type RescheduleTaskInput struct {
	UserID  ids.UserID
	TaskID  ids.TaskID
	DueDate time.Time
	Source  events.Source
}

func (uc *RescheduleTask) Execute(ctx context.Context, in RescheduleTaskInput) (TaskDTO, error) {
	if in.UserID.IsZero() || in.TaskID.IsZero() {
		return TaskDTO{}, fmt.Errorf("user id and task id are required")
	}
	if in.DueDate.IsZero() {
		return TaskDTO{}, fmt.Errorf("due date is required")
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
		from := task.DueDate
		if err := task.Reschedule(in.DueDate, now); err != nil {
			return err
		}
		if err := uc.store.Update(txCtx, task); err != nil {
			return err
		}
		if err := uc.events.Append(txCtx, events.Record{
			UserID:        task.UserID,
			AggregateType: "task",
			AggregateID:   task.ID.UUID(),
			EventType:     "TaskRescheduled",
			Payload: map[string]any{
				"from": from, "to": task.DueDate, "title": task.Title,
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
		return TaskDTO{}, fmt.Errorf("reschedule task: %w", err)
	}
	return result, nil
}
