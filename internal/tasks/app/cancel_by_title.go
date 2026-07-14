package app

import (
	"context"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type CancelTaskByTitle struct {
	store  TaskStore
	cancel *CancelTask
}

func NewCancelTaskByTitle(store TaskStore, cancel *CancelTask) *CancelTaskByTitle {
	return &CancelTaskByTitle{store: store, cancel: cancel}
}

func (uc *CancelTaskByTitle) Execute(ctx context.Context, userID ids.UserID, title string, source events.Source) (TaskDTO, error) {
	task, err := uc.store.FindOpenByTitle(ctx, userID, title)
	if err != nil {
		return TaskDTO{}, ErrTaskNotFound
	}
	return uc.cancel.Execute(ctx, CancelTaskInput{UserID: userID, TaskID: task.ID, Source: source})
}

type RescheduleTaskByTitle struct {
	store      TaskStore
	reschedule *RescheduleTask
}

func NewRescheduleTaskByTitle(store TaskStore, reschedule *RescheduleTask) *RescheduleTaskByTitle {
	return &RescheduleTaskByTitle{store: store, reschedule: reschedule}
}

func (uc *RescheduleTaskByTitle) Execute(ctx context.Context, userID ids.UserID, title string, dueDate time.Time, source events.Source) (TaskDTO, error) {
	task, err := uc.store.FindOpenByTitle(ctx, userID, title)
	if err != nil {
		return TaskDTO{}, ErrTaskNotFound
	}
	return uc.reschedule.Execute(ctx, RescheduleTaskInput{
		UserID: userID, TaskID: task.ID, DueDate: dueDate, Source: source,
	})
}
