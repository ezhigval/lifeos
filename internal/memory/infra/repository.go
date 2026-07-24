package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/memory/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Upsert(ctx context.Context, mem domain.Memory) (domain.Memory, error) {
	const q = `
		INSERT INTO user_memories (
			id, user_id, kind, key, value, confidence, source, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, kind, key) DO UPDATE SET
			value = EXCLUDED.value,
			confidence = EXCLUDED.confidence,
			source = EXCLUDED.source,
			updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, kind, key, value, confidence, source, created_at, updated_at
	`

	var out domain.Memory
	var kind string
	err := r.pool.QueryRow(ctx, q,
		mem.ID.UUID(),
		mem.UserID.UUID(),
		string(mem.Kind),
		mem.Key,
		mem.Value,
		mem.Confidence,
		mem.Source,
		mem.CreatedAt,
		mem.UpdatedAt,
	).Scan(
		&out.ID,
		&out.UserID,
		&kind,
		&out.Key,
		&out.Value,
		&out.Confidence,
		&out.Source,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return domain.Memory{}, fmt.Errorf("upsert user_memories: %w", err)
	}
	out.Kind = domain.Kind(kind)
	return out, nil
}

func (r *Repository) List(ctx context.Context, userID ids.UserID, limit int) ([]domain.Memory, error) {
	const q = `
		SELECT id, user_id, kind, key, value, confidence, source, created_at, updated_at
		FROM user_memories
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`
	return r.scanList(ctx, q, userID.UUID(), limit)
}

func (r *Repository) Recall(ctx context.Context, userID ids.UserID, query string, limit int) ([]domain.Memory, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return r.List(ctx, userID, limit)
	}
	pattern := "%" + escapeLike(query) + "%"
	const q = `
		SELECT id, user_id, kind, key, value, confidence, source, created_at, updated_at
		FROM user_memories
		WHERE user_id = $1
		  AND (key ILIKE $2 ESCAPE '\' OR value ILIKE $2 ESCAPE '\')
		ORDER BY updated_at DESC
		LIMIT $3
	`
	return r.scanList(ctx, q, userID.UUID(), pattern, limit)
}

func (r *Repository) Delete(ctx context.Context, userID ids.UserID, memoryID ids.MemoryID) error {
	const q = `
		DELETE FROM user_memories
		WHERE id = $1 AND user_id = $2
	`
	tag, err := r.pool.Exec(ctx, q, memoryID.UUID(), userID.UUID())
	if err != nil {
		return fmt.Errorf("delete user_memories: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *Repository) GetPrivacyFlags(ctx context.Context, userID ids.UserID) (domain.PrivacyFlags, error) {
	const q = `
		SELECT memory_enabled, learning_opt_in
		FROM user_settings
		WHERE user_id = $1
	`
	var flags domain.PrivacyFlags
	err := r.pool.QueryRow(ctx, q, userID.UUID()).Scan(&flags.MemoryEnabled, &flags.LearningOptIn)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PrivacyFlags{MemoryEnabled: true, LearningOptIn: false}, nil
	}
	if err != nil {
		return domain.PrivacyFlags{}, fmt.Errorf("get privacy flags: %w", err)
	}
	return flags, nil
}

func (r *Repository) SetPrivacyFlags(ctx context.Context, userID ids.UserID, flags domain.PrivacyFlags) error {
	const q = `
		INSERT INTO user_settings (user_id, memory_enabled, learning_opt_in)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			memory_enabled = EXCLUDED.memory_enabled,
			learning_opt_in = EXCLUDED.learning_opt_in,
			updated_at = now()
	`
	_, err := r.pool.Exec(ctx, q, userID.UUID(), flags.MemoryEnabled, flags.LearningOptIn)
	if err != nil {
		return fmt.Errorf("set privacy flags: %w", err)
	}
	return nil
}

func (r *Repository) scanList(ctx context.Context, q string, args ...any) ([]domain.Memory, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query user_memories: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Memory, 0)
	for rows.Next() {
		var mem domain.Memory
		var kind string
		if err := rows.Scan(
			&mem.ID,
			&mem.UserID,
			&kind,
			&mem.Key,
			&mem.Value,
			&mem.Confidence,
			&mem.Source,
			&mem.CreatedAt,
			&mem.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user_memories: %w", err)
		}
		mem.Kind = domain.Kind(kind)
		out = append(out, mem)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user_memories: %w", err)
	}
	return out, nil
}

func escapeLike(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
}
