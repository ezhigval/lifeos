package domain

import (
	"errors"
	"math"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

const maxSleepMinutes = 24 * 60

var ErrInvalidSleep = errors.New("sleep duration must be between 1 minute and 24 hours")

type SleepLog struct {
	ID              ids.SleepLogID
	UserID          ids.UserID
	DurationMinutes int32
	LoggedAt        time.Time
	CreatedAt       time.Time
}

func NewSleepLog(userID ids.UserID, durationMinutes int32, loggedAt, now time.Time) (SleepLog, error) {
	if durationMinutes <= 0 || durationMinutes > maxSleepMinutes {
		return SleepLog{}, ErrInvalidSleep
	}
	now = now.UTC()
	loggedAt = loggedAt.UTC()
	return SleepLog{
		ID:              ids.NewSleepLogID(),
		UserID:          userID,
		DurationMinutes: durationMinutes,
		LoggedAt:        loggedAt,
		CreatedAt:       now,
	}, nil
}

func DurationHours(minutes int32) float64 {
	return float64(minutes) / 60.0
}

func MinutesFromHours(hours float64) int32 {
	return int32(math.Round(hours * 60))
}
