package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByTelegramID(ctx context.Context, telegramID int64) (domain.User, error) {
	const q = `
		SELECT id, telegram_id, display_name, timezone, created_at
		FROM users
		WHERE telegram_id = $1
	`

	var (
		id          ids.UserID
		tgID        int64
		displayName string
		timezone    string
		createdAt   time.Time
	)

	err := r.pool.QueryRow(ctx, q, telegramID).Scan(
		&id, &tgID, &displayName, &timezone, &createdAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("query user: %w", err)
	}

	return domain.User{
		ID:          id,
		TelegramID:  tgID,
		DisplayName: displayName,
		Timezone:    timezone,
		CreatedAt:   createdAt,
	}, nil
}

func (r *Repository) Upsert(ctx context.Context, user domain.User) error {
	const q = `
		INSERT INTO users (id, telegram_id, display_name, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (telegram_id) DO UPDATE
		SET display_name = EXCLUDED.display_name,
		    timezone = EXCLUDED.timezone,
		    updated_at = now()
	`

	_, err := r.pool.Exec(ctx, q,
		user.ID.UUID(),
		user.TelegramID,
		user.DisplayName,
		user.Timezone,
		user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, userID ids.UserID) (domain.User, error) {
	const q = `
		SELECT id, telegram_id, display_name, timezone, created_at
		FROM users
		WHERE id = $1
	`

	var (
		id          ids.UserID
		tgID        int64
		displayName string
		timezone    string
		createdAt   time.Time
	)

	err := r.pool.QueryRow(ctx, q, userID.UUID()).Scan(
		&id, &tgID, &displayName, &timezone, &createdAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("query user: %w", err)
	}

	return domain.User{
		ID:          id,
		TelegramID:  tgID,
		DisplayName: displayName,
		Timezone:    timezone,
		CreatedAt:   createdAt,
	}, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]domain.User, error) {
	const q = `
		SELECT id, telegram_id, display_name, timezone, created_at
		FROM users
		ORDER BY created_at
	`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var (
			id          ids.UserID
			tgID        int64
			displayName string
			timezone    string
			createdAt   time.Time
		)
		if err := rows.Scan(&id, &tgID, &displayName, &timezone, &createdAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, domain.User{
			ID:          id,
			TelegramID:  tgID,
			DisplayName: displayName,
			Timezone:    timezone,
			CreatedAt:   createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return out, nil
}
