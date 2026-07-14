package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// Delete permanently wipes the user and all owned domain rows, then removes
// the users row. Explicit deletes avoid project_spheres ON DELETE RESTRICT and
// keep the wipe correct even when some CASCADE edges are missing.
func (r *Repository) Delete(ctx context.Context, userID ids.UserID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete user: %w", err)
	}
	defer tx.Rollback(ctx)

	uid := userID.UUID()
	stmts := []string{
		`DELETE FROM task_projects WHERE task_id IN (SELECT id FROM tasks WHERE user_id = $1)
		   OR project_id IN (SELECT id FROM projects WHERE user_id = $1)`,
		`DELETE FROM project_spheres WHERE project_id IN (SELECT id FROM projects WHERE user_id = $1)`,
		`DELETE FROM habit_logs WHERE habit_id IN (SELECT id FROM habits WHERE user_id = $1)`,
		`DELETE FROM planned_cashflows WHERE user_id = $1`,
		`DELETE FROM finance_transactions WHERE user_id = $1`,
		`DELETE FROM finance_categories WHERE user_id = $1`,
		`DELETE FROM debts WHERE user_id = $1`,
		`DELETE FROM tasks WHERE user_id = $1`,
		`DELETE FROM projects WHERE user_id = $1`,
		`DELETE FROM habits WHERE user_id = $1`,
		`DELETE FROM notes WHERE user_id = $1`,
		`DELETE FROM calendar_events WHERE user_id = $1`,
		`DELETE FROM career_contacts WHERE user_id = $1`,
		`DELETE FROM career_skills WHERE user_id = $1`,
		`DELETE FROM health_weight_logs WHERE user_id = $1`,
		`DELETE FROM health_step_logs WHERE user_id = $1`,
		`DELETE FROM health_sleep_logs WHERE user_id = $1`,
		`DELETE FROM day_availability WHERE user_id = $1`,
		`DELETE FROM scheduled_jobs WHERE user_id = $1`,
		`DELETE FROM domain_events WHERE user_id = $1`,
		`DELETE FROM telegram_sessions WHERE user_id = $1`,
		`DELETE FROM life_spheres WHERE user_id = $1`,
		`DELETE FROM user_settings WHERE user_id = $1`,
	}
	for _, q := range stmts {
		if _, err := tx.Exec(ctx, q, uid); err != nil {
			if isUndefinedTable(err) {
				continue
			}
			return fmt.Errorf("wipe user data: %w", err)
		}
	}
	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, uid)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete user: %w", err)
	}
	return nil
}

func isUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "does not exist")
}

// Silence unused import if pool helpers referenced elsewhere.
var _ = (*pgxpool.Pool)(nil)
