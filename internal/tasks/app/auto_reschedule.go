package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
)

type AutoRescheduleIncomplete struct {
	store      TaskStore
	events     EventLog
	transactor Transactor
	users      UserReader
	now        func() time.Time
}

func NewAutoRescheduleIncomplete(store TaskStore, events EventLog, transactor Transactor, users UserReader) *AutoRescheduleIncomplete {
	return &AutoRescheduleIncomplete{
		store: store, events: events, transactor: transactor, users: users,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type AutoRescheduleResult struct {
	Moved []TaskDTO
}

func (uc *AutoRescheduleIncomplete) Execute(ctx context.Context, userID ids.UserID, source events.Source) (AutoRescheduleResult, error) {
	if userID.IsZero() {
		return AutoRescheduleResult{}, fmt.Errorf("user id is required")
	}
	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return AutoRescheduleResult{}, err
	}
	today, err := timeutil.DateInTimezone(uc.now(), tz)
	if err != nil {
		return AutoRescheduleResult{}, err
	}
	tomorrow := today.Add(24 * time.Hour)
	now := uc.now()

	var moved []TaskDTO
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		tasks, err := uc.store.ListOpenDueOnOrBefore(txCtx, userID, today)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			from := task.DueDate
			if err := task.Reschedule(tomorrow, now); err != nil {
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
					"from": from, "to": task.DueDate, "title": task.Title, "auto": true,
				},
				Source:     source,
				OccurredAt: now,
			}); err != nil {
				return err
			}
			moved = append(moved, ToDTO(task))
		}
		return nil
	})
	if err != nil {
		return AutoRescheduleResult{}, fmt.Errorf("auto reschedule: %w", err)
	}
	return AutoRescheduleResult{Moved: moved}, nil
}
