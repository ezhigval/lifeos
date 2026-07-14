package app

import (
	"context"
	"fmt"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/settings/domain"
)

type GetSettings struct {
	store SettingsStore
}

func NewGetSettings(store SettingsStore) *GetSettings {
	return &GetSettings{store: store}
}

type SettingsDTO struct {
	MorningReviewAt domain.TimeOfDay
	EveningReviewAt domain.TimeOfDay
	WeeklyReviewAt  domain.TimeOfDay
	MonthlyReviewAt domain.TimeOfDay
	QuietHoursStart *domain.TimeOfDay
	QuietHoursEnd   *domain.TimeOfDay
	Language        string
}

func (uc *GetSettings) Execute(ctx context.Context, userID ids.UserID) (SettingsDTO, error) {
	if userID.IsZero() {
		return SettingsDTO{}, fmt.Errorf("user id is required")
	}
	s, err := uc.store.Get(ctx, userID)
	if err != nil {
		return SettingsDTO{}, fmt.Errorf("get settings: %w", err)
	}
	return SettingsDTO{
		MorningReviewAt: s.MorningReviewAt,
		EveningReviewAt: s.EveningReviewAt,
		WeeklyReviewAt:  s.WeeklyReviewAt,
		MonthlyReviewAt: s.MonthlyReviewAt,
		QuietHoursStart: s.QuietHoursStart,
		QuietHoursEnd:   s.QuietHoursEnd,
		Language:        s.Language,
	}, nil
}
