package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
)

type Repository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (domain.User, error)
	Upsert(ctx context.Context, user domain.User) error
}

type SeedUser struct {
	repo Repository
	now  func() time.Time
}

func NewSeedUser(repo Repository) *SeedUser {
	return &SeedUser{
		repo: repo,
		now:  time.Now,
	}
}

type SeedInput struct {
	TelegramID  int64
	DisplayName string
	Timezone    string
}

func (uc *SeedUser) Execute(ctx context.Context, in SeedInput) (domain.User, error) {
	user, err := domain.NewUser(in.TelegramID, in.DisplayName, in.Timezone, uc.now().UTC())
	if err != nil {
		return domain.User{}, err
	}

	if err := uc.repo.Upsert(ctx, user); err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}

	got, err := uc.repo.GetByTelegramID(ctx, in.TelegramID)
	if err != nil {
		return domain.User{}, fmt.Errorf("reload user: %w", err)
	}
	return got, nil
}
