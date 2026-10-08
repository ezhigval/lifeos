package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/app"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

func TestListCalendarTasks(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	userID := ids.NewUserID()
	now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	from := now
	to := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	open, err := domain.NewTask(userID, "plan week", domain.PriorityMedium, &due, now)
	if err != nil {
		t.Fatal(err)
	}
	done, err := domain.NewTask(userID, "ship notes", domain.PriorityMedium, &due, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := done.Complete(now); err != nil {
		t.Fatal(err)
	}
	laterDue := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	later, err := domain.NewTask(userID, "next quarter", domain.PriorityLow, &laterDue, now)
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range []domain.Task{open, done, later} {
		if err := store.Save(context.Background(), task); err != nil {
			t.Fatal(err)
		}
	}

	items, err := app.NewListCalendarTasks(store).Execute(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	got := map[ids.TaskID]domain.Status{}
	for _, item := range items {
		got[item.ID] = item.Status
	}
	if got[open.ID] != domain.StatusTodo || got[done.ID] != domain.StatusDone || len(got) != 2 {
		t.Fatalf("items = %+v", items)
	}

	if _, err := app.NewListCalendarTasks(store).Execute(context.Background(), ids.UserID{}, from, to); err == nil || !strings.Contains(err.Error(), "user id is required") {
		t.Fatalf("zero user err = %v", err)
	}
	if _, err := app.NewListCalendarTasks(store).Execute(context.Background(), userID, to, from); err == nil || !strings.Contains(err.Error(), "invalid date range") {
		t.Fatalf("range err = %v", err)
	}

	storeErr := errors.New("list failed")
	if _, err := app.NewListCalendarTasks(calendarListError{err: storeErr}).Execute(context.Background(), userID, from, to); !errors.Is(err, storeErr) {
		t.Fatalf("store err = %v", err)
	}
}

// calendarListError satisfies TaskStore so ListCalendarTasks can surface a store failure.
type calendarListError struct {
	*fakeStore
	err error
}

func (s calendarListError) ListAllDueBetween(context.Context, ids.UserID, time.Time, time.Time) ([]domain.Task, error) {
	return nil, s.err
}
