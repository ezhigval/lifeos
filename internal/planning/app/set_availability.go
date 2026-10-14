package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/planning/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
)

type AvailabilityStore interface {
	Upsert(ctx context.Context, a domain.DayAvailability) error
}

type UserTimezone interface {
	Timezone(ctx context.Context, userID ids.UserID) (string, error)
}

type SetDayAvailability struct {
	store AvailabilityStore
	users UserTimezone
	now   func() time.Time
}

func NewSetDayAvailability(store AvailabilityStore, users UserTimezone) *SetDayAvailability {
	return &SetDayAvailability{
		store: store,
		users: users,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

func (uc *SetDayAvailability) Execute(ctx context.Context, userID ids.UserID, hour, minute int) (string, error) {
	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return "", err
	}
	day, err := timeutil.DateInTimezone(uc.now(), tz)
	if err != nil {
		return "", err
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", fmt.Errorf("load timezone %q: %w", tz, err)
	}
	// hour:minute is the user's wall clock, not UTC.
	until := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, loc)
	a, err := domain.NewDayAvailability(userID, day, until)
	if err != nil {
		return "", err
	}
	if err := uc.store.Upsert(ctx, a); err != nil {
		return "", fmt.Errorf("set availability: %w", err)
	}
	return fmt.Sprintf("%02d:%02d", hour, minute), nil
}
