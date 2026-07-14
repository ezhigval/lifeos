package domain

import "context"

type Repository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (User, error)
	Upsert(ctx context.Context, user User) error
}
