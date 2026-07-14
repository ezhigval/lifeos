package app

import (
	"context"
	"fmt"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
)

type UserRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (domain.User, error)
}

type GetUserByTelegram struct {
	repo UserRepository
}

func NewGetUserByTelegram(repo UserRepository) *GetUserByTelegram {
	return &GetUserByTelegram{repo: repo}
}

func (uc *GetUserByTelegram) Execute(ctx context.Context, telegramID int64) (domain.User, error) {
	if telegramID <= 0 {
		return domain.User{}, fmt.Errorf("invalid telegram id")
	}
	return uc.repo.GetByTelegramID(ctx, telegramID)
}
