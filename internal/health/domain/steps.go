package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrInvalidSteps = errors.New("steps must be between 1 and 200000")

type StepLog struct {
	ID        ids.StepLogID
	UserID    ids.UserID
	Steps     int32
	LoggedAt  time.Time
	CreatedAt time.Time
}

func NewStepLog(userID ids.UserID, steps int32, loggedAt, now time.Time) (StepLog, error) {
	if steps <= 0 || steps > 200000 {
		return StepLog{}, ErrInvalidSteps
	}
	now = now.UTC()
	loggedAt = loggedAt.UTC()
	return StepLog{
		ID:        ids.NewStepLogID(),
		UserID:    userID,
		Steps:     steps,
		LoggedAt:  loggedAt,
		CreatedAt: now,
	}, nil
}
