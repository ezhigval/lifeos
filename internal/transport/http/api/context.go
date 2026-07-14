package api

import (
	"context"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type ctxKey int

const userIDKey ctxKey = 1

func WithUserID(ctx context.Context, userID ids.UserID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (ids.UserID, bool) {
	v, ok := ctx.Value(userIDKey).(ids.UserID)
	return v, ok && !v.IsZero()
}
