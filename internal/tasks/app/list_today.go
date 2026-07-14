package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
)

type UserReader interface {
	Timezone(ctx context.Context, userID ids.UserID) (string, error)
}

type ListTasksToday struct {
	store TaskStore
	users UserReader
	now   func() time.Time
}

func NewListTasksToday(store TaskStore, users UserReader) *ListTasksToday {
	return &ListTasksToday{
		store: store,
		users: users,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

func (uc *ListTasksToday) Execute(ctx context.Context, userID ids.UserID) ([]TaskDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}

	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load timezone: %w", err)
	}

	today, err := timeutil.DateInTimezone(uc.now(), tz)
	if err != nil {
		return nil, err
	}

	tasks, err := uc.store.ListByDueDate(ctx, userID, today)
	if err != nil {
		return nil, fmt.Errorf("list tasks today: %w", err)
	}

	return ToDTOs(tasks), nil
}
