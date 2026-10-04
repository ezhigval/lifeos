package domain

import (
	"errors"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrInvalidReviewTime = errors.New("review time is required")

type UserSettings struct {
	UserID           ids.UserID
	MorningReviewAt  TimeOfDay
	EveningReviewAt  TimeOfDay
	WeeklyReviewAt   TimeOfDay
	MonthlyReviewAt  TimeOfDay
	QuietHoursStart  *TimeOfDay
	QuietHoursEnd    *TimeOfDay
	Language         string
	MemoryEnabled    bool
	LearningOptIn    bool
	// HomeWidgets — настройки видимости блоков на главной (TASK-011 п.6).
	// Ключ виджета -> виден ли; отсутствие ключа означает «виден» (default-on).
	HomeWidgets map[string]bool
}

// Known home widget keys (TASK-011 п.6). Unknown keys are rejected by the API.
var HomeWidgetKeys = []string{
	"tasks", "habits", "finance", "notes", "calendar",
	"reminders", "health", "debts", "analytics",
}

func IsKnownHomeWidget(key string) bool {
	for _, k := range HomeWidgetKeys {
		if k == key {
			return true
		}
	}
	return false
}

// WidgetVisible reports whether a home widget should be rendered for these settings.
func (s UserSettings) WidgetVisible(key string) bool {
	if v, ok := s.HomeWidgets[key]; ok {
		return v
	}
	return true // default: visible
}

// TimeOfDay represents HH:MM in user's local context.
type TimeOfDay struct {
	Hour   int
	Minute int
}

func DefaultSettings(userID ids.UserID) UserSettings {
	return UserSettings{
		UserID:          userID,
		MorningReviewAt: TimeOfDay{Hour: 8, Minute: 0},
		EveningReviewAt: TimeOfDay{Hour: 21, Minute: 0},
		WeeklyReviewAt:  TimeOfDay{Hour: 10, Minute: 0},
		MonthlyReviewAt: TimeOfDay{Hour: 10, Minute: 0},
		Language:        "ru",
		MemoryEnabled:   true,
		LearningOptIn:   false,
	}
}

func (t TimeOfDay) Valid() bool {
	return t.Hour >= 0 && t.Hour < 24 && t.Minute >= 0 && t.Minute < 60
}
