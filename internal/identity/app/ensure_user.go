package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type UserUpsertRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (domain.User, error)
	Upsert(ctx context.Context, user domain.User) error
}

type SettingsEnsurer interface {
	Execute(ctx context.Context, userID ids.UserID) error
}

type EnsureUserByTelegram struct {
	repo      UserUpsertRepository
	settings  SettingsEnsurer
	onCreated func(ctx context.Context, user domain.User) error
	defaultTZ string
	now       func() time.Time
}

type EnsureUserInput struct {
	TelegramID  int64
	DisplayName string
}

func NewEnsureUserByTelegram(
	repo UserUpsertRepository,
	settings SettingsEnsurer,
	defaultTZ string,
	onCreated func(ctx context.Context, user domain.User) error,
) *EnsureUserByTelegram {
	return &EnsureUserByTelegram{
		repo:      repo,
		settings:  settings,
		onCreated: onCreated,
		defaultTZ: defaultTZ,
		now:       time.Now,
	}
}

func (uc *EnsureUserByTelegram) Execute(ctx context.Context, in EnsureUserInput) (domain.User, error) {
	if in.TelegramID <= 0 {
		return domain.User{}, fmt.Errorf("invalid telegram id")
	}

	user, err := uc.repo.GetByTelegramID(ctx, in.TelegramID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, err
	}

	displayName := in.DisplayName
	if displayName == "" {
		displayName = fmt.Sprintf("User %d", in.TelegramID)
	}

	user, err = domain.NewUser(in.TelegramID, displayName, uc.defaultTZ, uc.now().UTC())
	if err != nil {
		return domain.User{}, err
	}
	if err := uc.repo.Upsert(ctx, user); err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	if uc.settings != nil {
		if err := uc.settings.Execute(ctx, user.ID); err != nil {
			return domain.User{}, fmt.Errorf("ensure settings: %w", err)
		}
	}
	if uc.onCreated != nil {
		if err := uc.onCreated(ctx, user); err != nil {
			return domain.User{}, fmt.Errorf("on user created: %w", err)
		}
	}
	return user, nil
}
