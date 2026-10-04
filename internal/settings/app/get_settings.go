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
	HomeWidgets     map[string]bool
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
		HomeWidgets:     s.HomeWidgets,
	}, nil
}

// UpdateHomeWidgets toggles which blocks are shown on the home screen (TASK-011 п.6).
type UpdateHomeWidgets struct {
	store SettingsStore
}

func NewUpdateHomeWidgets(store SettingsStore) *UpdateHomeWidgets {
	return &UpdateHomeWidgets{store: store}
}

func (uc *UpdateHomeWidgets) Execute(ctx context.Context, userID ids.UserID, widgets map[string]bool) (map[string]bool, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	for k := range widgets {
		if !domain.IsKnownHomeWidget(k) {
			return nil, fmt.Errorf("unknown widget key %q", k)
		}
	}
	current, err := uc.store.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	merged := make(map[string]bool, len(current.HomeWidgets)+len(widgets))
	for k, v := range current.HomeWidgets {
		merged[k] = v
	}
	for k, v := range widgets {
		merged[k] = v
	}
	if err := uc.store.UpdateHomeWidgets(ctx, userID, merged); err != nil {
		return nil, err
	}
	return merged, nil
}
