package app

import (
	"context"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/habits/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type habitMem struct {
	habit domain.Habit
}

func (s *habitMem) Save(context.Context, domain.Habit) error { return nil }
func (s *habitMem) Update(_ context.Context, habit domain.Habit) (bool, error) {
	s.habit = habit
	return true, nil
}
func (s *habitMem) Delete(context.Context, ids.UserID, ids.HabitID) (bool, error) {
	return true, nil
}
func (s *habitMem) GetByID(context.Context, ids.UserID, ids.HabitID) (domain.Habit, error) {
	return s.habit, nil
}
func (s *habitMem) FindByName(context.Context, ids.UserID, string) (domain.Habit, error) {
	return s.habit, nil
}
func (s *habitMem) ListWithToday(context.Context, ids.UserID, time.Time) ([]HabitDayRow, error) {
	return nil, nil
}

type logMem struct{ upserts int }

func (s *logMem) Upsert(context.Context, domain.HabitLog) error {
	s.upserts++
	return nil
}
func (s *logMem) ListSince(context.Context, ids.HabitID, time.Time) ([]domain.HabitLog, error) {
	return nil, nil
}

type habitEvents struct{}

func (habitEvents) Append(context.Context, events.Record) error { return nil }

type habitTx struct{}

func (habitTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type habitTZ struct{}

func (habitTZ) Timezone(context.Context, ids.UserID) (string, error) { return "UTC", nil }

func TestUpdateHabitKeepsOmittedDates(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	habit, err := domain.NewHabitWithDates(user, "вода", domain.FrequencyDaily, &start, &end, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	store := &habitMem{habit: habit}
	uc := NewUpdateHabit(store, habitEvents{}, habitTx{})

	dto, err := uc.Execute(context.Background(), UpdateHabitInput{
		UserID:  user,
		HabitID: habit.ID,
		Name:    "вода 2л",
		Source:  events.SourceHTTP,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dto.Name != "вода 2л" {
		t.Fatalf("name %q", dto.Name)
	}
	if store.habit.StartDate == nil || !store.habit.StartDate.Equal(start) {
		t.Fatalf("start wiped: %v", store.habit.StartDate)
	}
	if store.habit.EndDate == nil || !store.habit.EndDate.Equal(end) {
		t.Fatalf("end wiped: %v", store.habit.EndDate)
	}
}

func TestTrackHabitRejectsInactiveWindow(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	end := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	habit, err := domain.NewHabitWithDates(user, "старая", domain.FrequencyDaily, nil, &end, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	store := &habitMem{habit: habit}
	logs := &logMem{}
	uc := NewTrackHabit(store, logs, habitEvents{}, habitTx{}, habitTZ{})
	uc.now = func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) }

	_, err = uc.Execute(context.Background(), TrackHabitInput{
		UserID: user,
		Name:   "старая",
		Source: events.SourceHTTP,
	})
	if err != domain.ErrInactive {
		t.Fatalf("err %v", err)
	}
	if logs.upserts != 0 {
		t.Fatalf("logged %d times", logs.upserts)
	}
}
