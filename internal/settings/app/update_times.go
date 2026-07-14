package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/settings/domain"
)

type SettingsStore interface {
	Get(ctx context.Context, userID ids.UserID) (domain.UserSettings, error)
	UpdateMorningReview(ctx context.Context, userID ids.UserID, at domain.TimeOfDay) error
	UpdateEveningReview(ctx context.Context, userID ids.UserID, at domain.TimeOfDay) error
	UpdateQuietHours(ctx context.Context, userID ids.UserID, start, end domain.TimeOfDay) error
}

type ReviewRescheduler interface {
	RescheduleReview(ctx context.Context, userID ids.UserID, jobType string, runAt time.Time) error
}

type UpdateMorningReview struct {
	store       SettingsStore
	rescheduler ReviewRescheduler
	reviewAt    func(ref time.Time, tz string, tod domain.TimeOfDay) (time.Time, error)
	tz          func(ctx context.Context, userID ids.UserID) (string, error)
	now         func() time.Time
}

func NewUpdateMorningReview(
	store SettingsStore,
	rescheduler ReviewRescheduler,
	tz func(ctx context.Context, userID ids.UserID) (string, error),
	reviewAt func(ref time.Time, tz string, tod domain.TimeOfDay) (time.Time, error),
) *UpdateMorningReview {
	return &UpdateMorningReview{
		store:       store,
		rescheduler: rescheduler,
		tz:          tz,
		reviewAt:    reviewAt,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (uc *UpdateMorningReview) Execute(ctx context.Context, userID ids.UserID, at domain.TimeOfDay) (domain.TimeOfDay, error) {
	if userID.IsZero() || !at.Valid() {
		return domain.TimeOfDay{}, fmt.Errorf("invalid input")
	}
	if err := uc.store.UpdateMorningReview(ctx, userID, at); err != nil {
		return domain.TimeOfDay{}, err
	}
	if err := uc.reschedule(ctx, userID, "morning_review", at); err != nil {
		return domain.TimeOfDay{}, err
	}
	return at, nil
}

type UpdateEveningReview struct {
	*UpdateMorningReview
}

func NewUpdateEveningReview(
	store SettingsStore,
	rescheduler ReviewRescheduler,
	tz func(ctx context.Context, userID ids.UserID) (string, error),
	reviewAt func(ref time.Time, tz string, tod domain.TimeOfDay) (time.Time, error),
) *UpdateEveningReview {
	return &UpdateEveningReview{UpdateMorningReview: NewUpdateMorningReview(store, rescheduler, tz, reviewAt)}
}

func (uc *UpdateEveningReview) Execute(ctx context.Context, userID ids.UserID, at domain.TimeOfDay) (domain.TimeOfDay, error) {
	if userID.IsZero() || !at.Valid() {
		return domain.TimeOfDay{}, fmt.Errorf("invalid input")
	}
	if err := uc.store.UpdateEveningReview(ctx, userID, at); err != nil {
		return domain.TimeOfDay{}, err
	}
	if err := uc.reschedule(ctx, userID, "evening_review", at); err != nil {
		return domain.TimeOfDay{}, err
	}
	return at, nil
}

func (uc *UpdateMorningReview) reschedule(ctx context.Context, userID ids.UserID, jobType string, at domain.TimeOfDay) error {
	tz, err := uc.tz(ctx, userID)
	if err != nil {
		return err
	}
	runAt, err := uc.reviewAt(uc.now(), tz, at)
	if err != nil {
		return err
	}
	return uc.rescheduler.RescheduleReview(ctx, userID, jobType, runAt)
}

type UpdateQuietHours struct {
	store SettingsStore
}

func NewUpdateQuietHours(store SettingsStore) *UpdateQuietHours {
	return &UpdateQuietHours{store: store}
}

func (uc *UpdateQuietHours) Execute(ctx context.Context, userID ids.UserID, start, end domain.TimeOfDay) error {
	if userID.IsZero() || !start.Valid() || !end.Valid() {
		return fmt.Errorf("invalid input")
	}
	return uc.store.UpdateQuietHours(ctx, userID, start, end)
}
