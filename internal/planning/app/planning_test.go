package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/planning/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type tzStub struct{ tz string }

func (s tzStub) Timezone(context.Context, ids.UserID) (string, error) { return s.tz, nil }

type availMem struct{ last domain.DayAvailability }

func (s *availMem) Upsert(_ context.Context, a domain.DayAvailability) error {
	s.last = a
	return nil
}

func TestSetDayAvailabilityUsesUserTimezone(t *testing.T) {
	t.Parallel()
	store := &availMem{}
	uc := NewSetDayAvailability(store, tzStub{tz: "Europe/Moscow"})
	uc.now = func() time.Time { return time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC) }

	got, err := uc.Execute(context.Background(), ids.NewUserID(), 18, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != "18:00" {
		t.Fatalf("label %q", got)
	}
	// 18:00 in Moscow is 15:00 UTC, not 18:00 UTC.
	if h := store.last.AvailableUntil.UTC().Hour(); h != 15 {
		t.Fatalf("until UTC hour %d, want 15 (%s)", h, store.last.AvailableUntil.UTC())
	}
}

type taskMem struct {
	tasks []taskdomain.Task
}

func (s *taskMem) ListByDueDate(_ context.Context, _ ids.UserID, due time.Time) ([]taskdomain.Task, error) {
	var out []taskdomain.Task
	for _, task := range s.tasks {
		if task.DueDate != nil && task.DueDate.Equal(due) {
			out = append(out, task)
		}
	}
	return out, nil
}

func (s *taskMem) ListOpenDueOnOrBefore(_ context.Context, _ ids.UserID, due time.Time) ([]taskdomain.Task, error) {
	var out []taskdomain.Task
	for _, task := range s.tasks {
		if task.Status == taskdomain.StatusDone || task.Status == taskdomain.StatusCancelled {
			continue
		}
		if task.DueDate != nil && !task.DueDate.After(due) {
			out = append(out, task)
		}
	}
	return out, nil
}

func (s *taskMem) Update(_ context.Context, task taskdomain.Task) error {
	for i := range s.tasks {
		if s.tasks[i].ID == task.ID {
			s.tasks[i] = task
			return nil
		}
	}
	return nil
}

func TestTriageSkipsClosedTasks(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	now := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	today := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	open, err := taskdomain.NewTask(user, "открытая", taskdomain.PriorityLow, &today, now)
	if err != nil {
		t.Fatal(err)
	}
	done, err := taskdomain.NewTask(user, "сделанная", taskdomain.PriorityLow, &today, now)
	if err != nil {
		t.Fatal(err)
	}
	done.Status = taskdomain.StatusDone
	store := &taskMem{tasks: []taskdomain.Task{open, done}}
	uc := NewTriageOverloadedDay(store, tzStub{tz: "UTC"})
	uc.now = func() time.Time { return now }

	text, low, err := uc.Propose(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}
	if len(low) != 1 || low[0] != open.ID {
		t.Fatalf("low ids %v", low)
	}
	if strings.Contains(text, "сделанная") {
		t.Fatalf("closed task leaked into triage: %s", text)
	}
}

func TestRescheduleMovesOverdueOpenTasks(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	now := time.Date(2026, 3, 2, 8, 0, 0, 0, time.UTC)
	today := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	overdue, err := taskdomain.NewTask(user, "просрочена", taskdomain.PriorityHigh, &yesterday, now)
	if err != nil {
		t.Fatal(err)
	}
	doneToday, err := taskdomain.NewTask(user, "закрыта", taskdomain.PriorityLow, &today, now)
	if err != nil {
		t.Fatal(err)
	}
	doneToday.Status = taskdomain.StatusDone
	store := &taskMem{tasks: []taskdomain.Task{overdue, doneToday}}
	uc := NewRescheduleTasks(store, tzStub{tz: "UTC"})
	uc.now = func() time.Time { return now }

	n, err := uc.Execute(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("moved %d", n)
	}
	if store.tasks[0].DueDate == nil || !store.tasks[0].DueDate.Equal(today.AddDate(0, 0, 1)) {
		t.Fatalf("overdue due %v", store.tasks[0].DueDate)
	}
	if store.tasks[1].DueDate == nil || !store.tasks[1].DueDate.Equal(today) {
		t.Fatalf("done task moved: %v", store.tasks[1].DueDate)
	}
}
