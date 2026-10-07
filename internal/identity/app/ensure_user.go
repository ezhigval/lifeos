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
	// Username is the public Telegram nick without @. Empty leaves the stored nick alone.
	Username string
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
		uc.rememberUsername(ctx, user.ID, in.Username)
		user.TelegramUsername = normalizeStoredUsername(in.Username, user.TelegramUsername)
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
	uc.rememberUsername(ctx, user.ID, in.Username)
	user.TelegramUsername = normalizeStoredUsername(in.Username, "")
	return user, nil
}

type usernameRecorder interface {
	RememberUsername(ctx context.Context, userID ids.UserID, username string) error
}

func (uc *EnsureUserByTelegram) rememberUsername(ctx context.Context, userID ids.UserID, raw string) {
	username, ok := NormalizeTelegramUsername(raw)
	if !ok {
		return
	}
	rec, ok := uc.repo.(usernameRecorder)
	if !ok {
		return
	}
	// A nick collision must not fail /start. The next login lookup still works
	// once RememberUsername can move the nick.
	_ = rec.RememberUsername(ctx, userID, username)
}

func normalizeStoredUsername(raw, existing string) string {
	username, ok := NormalizeTelegramUsername(raw)
	if !ok {
		return existing
	}
	return username
}
