package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type TimezoneReader struct {
	pool *pgxpool.Pool
}

func NewTimezoneReader(pool *pgxpool.Pool) *TimezoneReader {
	return &TimezoneReader{pool: pool}
}

func (r *TimezoneReader) Timezone(ctx context.Context, userID ids.UserID) (string, error) {
	const q = `SELECT timezone FROM users WHERE id = $1`

	var timezone string
	err := r.pool.QueryRow(ctx, q, userID.UUID()).Scan(&timezone)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("user not found")
	}
	if err != nil {
		return "", fmt.Errorf("query timezone: %w", err)
	}
	return timezone, nil
}
