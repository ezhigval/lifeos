package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/habits/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
)

type HabitStore interface {
	Save(ctx context.Context, habit domain.Habit) error
	GetByID(ctx context.Context, userID ids.UserID, habitID ids.HabitID) (domain.Habit, error)
	FindByName(ctx context.Context, userID ids.UserID, name string) (domain.Habit, error)
	ListWithToday(ctx context.Context, userID ids.UserID, today time.Time) ([]HabitDayRow, error)
}

type HabitLogStore interface {
	Upsert(ctx context.Context, log domain.HabitLog) error
	ListSince(ctx context.Context, habitID ids.HabitID, since time.Time) ([]domain.HabitLog, error)
}

type HabitDayRow struct {
	Habit          domain.Habit
	TodayCompleted bool
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserReader interface {
	Timezone(ctx context.Context, userID ids.UserID) (string, error)
}

type HabitDTO struct {
	ID        ids.HabitID
	Name      string
	Frequency domain.Frequency
}

type HabitDayDTO struct {
	ID             ids.HabitID
	Name           string
	TodayCompleted bool
	Streak         int
}

type CreateHabit struct {
	habits     HabitStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreateHabit(habits HabitStore, events EventLog, transactor Transactor) *CreateHabit {
	return &CreateHabit{
		habits: habits, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateHabitInput struct {
	UserID ids.UserID
	Name   string
	Source events.Source
}

func (uc *CreateHabit) Execute(ctx context.Context, in CreateHabitInput) (HabitDTO, error) {
	if in.UserID.IsZero() {
		return HabitDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	habit, err := domain.NewHabit(in.UserID, in.Name, domain.FrequencyDaily, now)
	if err != nil {
		return HabitDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.habits.Save(txCtx, habit); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        habit.UserID,
			AggregateType: "habit",
			AggregateID:   habit.ID.UUID(),
			EventType:     "HabitCreated",
			Payload:       map[string]any{"name": habit.Name},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return HabitDTO{}, fmt.Errorf("create habit: %w", err)
	}
	return HabitDTO{ID: habit.ID, Name: habit.Name, Frequency: habit.Frequency}, nil
}

type TrackHabit struct {
	habits     HabitStore
	logs       HabitLogStore
	events     EventLog
	transactor Transactor
	users      UserReader
	now        func() time.Time
}

func NewTrackHabit(habits HabitStore, logs HabitLogStore, events EventLog, transactor Transactor, users UserReader) *TrackHabit {
	return &TrackHabit{
		habits: habits, logs: logs, events: events, transactor: transactor, users: users,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type TrackHabitInput struct {
	UserID ids.UserID
	Name   string
	Source events.Source
}

type TrackHabitResult struct {
	Name   string
	Streak int
}

func (uc *TrackHabit) Execute(ctx context.Context, in TrackHabitInput) (TrackHabitResult, error) {
	if in.UserID.IsZero() {
		return TrackHabitResult{}, fmt.Errorf("user id is required")
	}
	habit, err := uc.habits.FindByName(ctx, in.UserID, in.Name)
	if errors.Is(err, domain.ErrNotFound) {
		return TrackHabitResult{}, err
	}
	if err != nil {
		return TrackHabitResult{}, err
	}
	return uc.track(ctx, habit, in.Source)
}

func (uc *TrackHabit) ExecuteByID(ctx context.Context, userID ids.UserID, habitID ids.HabitID, source events.Source) (TrackHabitResult, error) {
	if userID.IsZero() || habitID.IsZero() {
		return TrackHabitResult{}, fmt.Errorf("user id and habit id are required")
	}
	habit, err := uc.habits.GetByID(ctx, userID, habitID)
	if errors.Is(err, domain.ErrNotFound) {
		return TrackHabitResult{}, err
	}
	if err != nil {
		return TrackHabitResult{}, err
	}
	return uc.track(ctx, habit, source)
}

func (uc *TrackHabit) track(ctx context.Context, habit domain.Habit, source events.Source) (TrackHabitResult, error) {
	now := uc.now()
	today, err := uc.today(ctx, habit.UserID, now)
	if err != nil {
		return TrackHabitResult{}, err
	}
	log := domain.NewHabitLog(habit.ID, today, now)

	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.logs.Upsert(txCtx, log); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        habit.UserID,
			AggregateType: "habit",
			AggregateID:   habit.ID.UUID(),
			EventType:     "HabitTracked",
			Payload:       map[string]any{"name": habit.Name, "log_date": today},
			Source:        source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return TrackHabitResult{}, fmt.Errorf("track habit: %w", err)
	}

	streak, err := uc.streak(ctx, habit.ID, today)
	if err != nil {
		streak = 1
	}
	return TrackHabitResult{Name: habit.Name, Streak: streak}, nil
}

func (uc *TrackHabit) today(ctx context.Context, userID ids.UserID, now time.Time) (time.Time, error) {
	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return time.Time{}, err
	}
	return timeutil.DateInTimezone(now, tz)
}

func (uc *TrackHabit) streak(ctx context.Context, habitID ids.HabitID, today time.Time) (int, error) {
	since := today.AddDate(0, 0, -400)
	logs, err := uc.logs.ListSince(ctx, habitID, since)
	if err != nil {
		return 0, err
	}
	dates := make([]time.Time, 0, len(logs))
	for _, log := range logs {
		if log.Completed {
			dates = append(dates, log.LogDate)
		}
	}
	return domain.Streak(dates, today), nil
}

type ListHabitsToday struct {
	habits HabitStore
	logs   HabitLogStore
	users  UserReader
	now    func() time.Time
}

func NewListHabitsToday(habits HabitStore, logs HabitLogStore, users UserReader) *ListHabitsToday {
	return &ListHabitsToday{
		habits: habits, logs: logs, users: users,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (uc *ListHabitsToday) Execute(ctx context.Context, userID ids.UserID) ([]HabitDayDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return nil, err
	}
	today, err := timeutil.DateInTimezone(uc.now(), tz)
	if err != nil {
		return nil, err
	}
	rows, err := uc.habits.ListWithToday(ctx, userID, today)
	if err != nil {
		return nil, fmt.Errorf("list habits: %w", err)
	}
	out := make([]HabitDayDTO, 0, len(rows))
	for _, row := range rows {
		streak, _ := uc.streakFor(ctx, row.Habit.ID, today)
		out = append(out, HabitDayDTO{
			ID:             row.Habit.ID,
			Name:           row.Habit.Name,
			TodayCompleted: row.TodayCompleted,
			Streak:         streak,
		})
	}
	return out, nil
}

func (uc *ListHabitsToday) streakFor(ctx context.Context, habitID ids.HabitID, today time.Time) (int, error) {
	since := today.AddDate(0, 0, -400)
	logs, err := uc.logs.ListSince(ctx, habitID, since)
	if err != nil {
		return 0, err
	}
	dates := make([]time.Time, 0, len(logs))
	for _, log := range logs {
		if log.Completed {
			dates = append(dates, log.LogDate)
		}
	}
	return domain.Streak(dates, today), nil
}
