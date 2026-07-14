package app

import (
	"context"
	"fmt"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type UserByIDRepository interface {
	GetByID(ctx context.Context, userID ids.UserID) (domain.User, error)
}

type GetUserByID struct {
	repo UserByIDRepository
}

func NewGetUserByID(repo UserByIDRepository) *GetUserByID {
	return &GetUserByID{repo: repo}
}

func (uc *GetUserByID) Execute(ctx context.Context, userID ids.UserID) (domain.User, error) {
	if userID.IsZero() {
		return domain.User{}, fmt.Errorf("invalid user id")
	}
	return uc.repo.GetByID(ctx, userID)
}
