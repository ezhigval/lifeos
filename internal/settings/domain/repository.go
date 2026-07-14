package domain

import (
	"context"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type Repository interface {
	EnsureDefaults(ctx context.Context, userID ids.UserID) error
}
