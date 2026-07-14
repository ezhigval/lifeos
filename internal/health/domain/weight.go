package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrInvalidWeight = errors.New("weight must be between 20 and 500 kg")
	ErrNotFound      = errors.New("weight log not found")
)

type WeightLog struct {
	ID        ids.WeightLogID
	UserID    ids.UserID
	WeightKg  float64
	LoggedAt  time.Time
	CreatedAt time.Time
}

func NewWeightLog(userID ids.UserID, weightKg float64, loggedAt, now time.Time) (WeightLog, error) {
	if weightKg < 20 || weightKg > 500 {
		return WeightLog{}, ErrInvalidWeight
	}
	now = now.UTC()
	loggedAt = loggedAt.UTC()
	return WeightLog{
		ID:        ids.NewWeightLogID(),
		UserID:    userID,
		WeightKg:  weightKg,
		LoggedAt:  loggedAt,
		CreatedAt: now,
	}, nil
}
