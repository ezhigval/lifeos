package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/identity/domain"
)

// LoginCodes stores one-time web login challenges. Raw digits never land here.
type LoginCodes struct {
	pool *pgxpool.Pool
}

func NewLoginCodes(pool *pgxpool.Pool) *LoginCodes {
	return &LoginCodes{pool: pool}
}

func (r *LoginCodes) CountSince(ctx context.Context, username string, since time.Time) (int, error) {
	const q = `
		SELECT count(*)
		FROM login_codes
		WHERE username = $1 AND created_at >= $2
	`
	var n int
	if err := r.pool.QueryRow(ctx, q, username, since).Scan(&n); err != nil {
		return 0, fmt.Errorf("count login codes: %w", err)
	}
	return n, nil
}

func (r *LoginCodes) LatestActive(ctx context.Context, username string, now time.Time) (app.LoginCode, error) {
	const q = `
		SELECT id, telegram_id, username, code_hash, expires_at, attempts, created_at
		FROM login_codes
		WHERE username = $1 AND expires_at > $2 AND attempts < 5
		ORDER BY created_at DESC
		LIMIT 1
	`
	var row app.LoginCode
	err := r.pool.QueryRow(ctx, q, username, now).Scan(
		&row.ID, &row.TelegramID, &row.Username, &row.CodeHash, &row.ExpiresAt, &row.Attempts, &row.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.LoginCode{}, domain.ErrNotFound
	}
	if err != nil {
		return app.LoginCode{}, fmt.Errorf("latest login code: %w", err)
	}
	return row, nil
}

func (r *LoginCodes) Insert(ctx context.Context, row app.LoginCode) error {
	const q = `
		INSERT INTO login_codes (id, telegram_id, username, code_hash, expires_at, attempts, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, q, row.ID, row.TelegramID, row.Username, row.CodeHash, row.ExpiresAt, row.Attempts, row.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert login code: %w", err)
	}
	return nil
}

func (r *LoginCodes) AddAttempt(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE login_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("login code attempt: %w", err)
	}
	return nil
}

func (r *LoginCodes) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM login_codes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete login code: %w", err)
	}
	return nil
}

func (r *LoginCodes) DeleteOlderThan(ctx context.Context, before time.Time) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM login_codes WHERE created_at < $1`, before)
	if err != nil {
		return fmt.Errorf("prune login codes: %w", err)
	}
	return nil
}
