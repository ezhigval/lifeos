package app

import (
	"context"
	"fmt"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/settings/domain"
)

type Repository interface {
	EnsureDefaults(ctx context.Context, userID ids.UserID) error
}

type EnsureDefaults struct {
	repo Repository
}

func NewEnsureDefaults(repo Repository) *EnsureDefaults {
	return &EnsureDefaults{repo: repo}
}

func (uc *EnsureDefaults) Execute(ctx context.Context, userID ids.UserID) error {
	if userID.IsZero() {
		return fmt.Errorf("user id is required")
	}
	if err := uc.repo.EnsureDefaults(ctx, userID); err != nil {
		return fmt.Errorf("ensure settings: %w", err)
	}
	return nil
}

// DefaultSettings returns the domain defaults for tests and docs.
func DefaultSettings(userID ids.UserID) domain.UserSettings {
	return domain.DefaultSettings(userID)
}
