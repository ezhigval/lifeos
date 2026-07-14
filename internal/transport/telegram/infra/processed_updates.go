package infra

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
)

type ProcessedUpdates struct {
	pool *pgxpool.Pool
}

func NewProcessedUpdates(pool *pgxpool.Pool) *ProcessedUpdates {
	return &ProcessedUpdates{pool: pool}
}

func (p *ProcessedUpdates) Seen(ctx context.Context, updateID int64) (bool, error) {
	return db.New(p.pool).ProcessedUpdateExists(ctx, updateID)
}

func (p *ProcessedUpdates) Mark(ctx context.Context, updateID int64) error {
	return db.New(p.pool).InsertProcessedUpdate(ctx, updateID)
}
