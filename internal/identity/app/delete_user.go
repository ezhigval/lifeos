package app

import (
	"context"
	"fmt"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type UserDeleteRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (domain.User, error)
	Delete(ctx context.Context, userID ids.UserID) error
}

// DeleteUser permanently removes a user and cascaded domain data.
type DeleteUser struct {
	repo UserDeleteRepository
}

func NewDeleteUser(repo UserDeleteRepository) *DeleteUser {
	return &DeleteUser{repo: repo}
}

func (uc *DeleteUser) Execute(ctx context.Context, telegramID int64) (domain.User, error) {
	if telegramID <= 0 {
		return domain.User{}, fmt.Errorf("invalid telegram id")
	}
	user, err := uc.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return domain.User{}, err
	}
	if err := uc.repo.Delete(ctx, user.ID); err != nil {
		return domain.User{}, fmt.Errorf("delete user: %w", err)
	}
	return user, nil
}

func (uc *DeleteUser) Lookup(ctx context.Context, telegramID int64) (domain.User, error) {
	if telegramID <= 0 {
		return domain.User{}, fmt.Errorf("invalid telegram id")
	}
	return uc.repo.GetByTelegramID(ctx, telegramID)
}
