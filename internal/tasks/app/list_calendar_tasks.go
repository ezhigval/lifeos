package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

// ListCalendarTasks returns ALL tasks (open and completed) with due_date in
// [from, to) — used by the calendar month/week/day views so users see both
// pending and done work on their timeline.
type ListCalendarTasks struct {
	store TaskStore
}

func NewListCalendarTasks(store TaskStore) *ListCalendarTasks {
	return &ListCalendarTasks{store: store}
}

func (uc *ListCalendarTasks) Execute(ctx context.Context, userID ids.UserID, from, to time.Time) ([]TaskDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("invalid date range")
	}
	items, err := uc.store.ListAllDueBetween(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	return ToDTOs(items), nil
}
