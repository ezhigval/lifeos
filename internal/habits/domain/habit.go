package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyName        = errors.New("habit name is required")
	ErrInvalidFrequency = errors.New("invalid habit frequency")
	ErrInvalidDeadline  = errors.New("habit deadline ends before it starts")
	ErrNotFound         = errors.New("habit not found")
)

type Frequency string

const FrequencyDaily Frequency = "daily"

func (f Frequency) Valid() bool {
	return f == FrequencyDaily
}

type Habit struct {
	ID        ids.HabitID
	UserID    ids.UserID
	Name      string
	Frequency Frequency
	StartDate *time.Time // optional; nil = from creation day
	EndDate   *time.Time // optional deadline ("срок"); nil = open-ended
	CreatedAt time.Time
}

// Active reports whether the habit applies on the given day (deadline window).
func (h Habit) Active(day time.Time) bool {
	if h.StartDate != nil && day.Before(*h.StartDate) {
		return false
	}
	if h.EndDate != nil && day.After(*h.EndDate) {
		return false
	}
	return true
}

func NewHabit(userID ids.UserID, name string, frequency Frequency, now time.Time) (Habit, error) {
	return NewHabitWithDates(userID, name, frequency, nil, nil, now)
}

func NewHabitWithDates(userID ids.UserID, name string, frequency Frequency, startDate, endDate *time.Time, now time.Time) (Habit, error) {
	if name == "" {
		return Habit{}, ErrEmptyName
	}
	if frequency == "" {
		frequency = FrequencyDaily
	}
	if !frequency.Valid() {
		return Habit{}, ErrInvalidFrequency
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return Habit{}, ErrInvalidDeadline
	}
	day := func(t *time.Time) *time.Time {
		if t == nil {
			return nil
		}
		d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		return &d
	}
	return Habit{
		ID:        ids.NewHabitID(),
		UserID:    userID,
		Name:      name,
		Frequency: frequency,
		StartDate: day(startDate),
		EndDate:   day(endDate),
		CreatedAt: now.UTC(),
	}, nil
}

type HabitLog struct {
	ID        ids.HabitLogID
	HabitID   ids.HabitID
	LogDate   time.Time
	Completed bool
	CreatedAt time.Time
}

func NewHabitLog(habitID ids.HabitID, logDate time.Time, now time.Time) HabitLog {
	day := time.Date(logDate.Year(), logDate.Month(), logDate.Day(), 0, 0, 0, 0, time.UTC)
	return HabitLog{
		ID:        ids.NewHabitLogID(),
		HabitID:   habitID,
		LogDate:   day,
		Completed: true,
		CreatedAt: now.UTC(),
	}
}

// Streak counts consecutive completed days ending at the latest completed day.
func Streak(completedDates []time.Time, today time.Time) int {
	if len(completedDates) == 0 {
		return 0
	}
	set := make(map[string]struct{}, len(completedDates))
	for _, d := range completedDates {
		key := d.Format("2006-01-02")
		set[key] = struct{}{}
	}

	check := today
	if _, ok := set[check.Format("2006-01-02")]; !ok {
		check = check.AddDate(0, 0, -1)
	}

	streak := 0
	for {
		if _, ok := set[check.Format("2006-01-02")]; !ok {
			break
		}
		streak++
		check = check.AddDate(0, 0, -1)
	}
	return streak
}
