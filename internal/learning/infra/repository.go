package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/learning/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Insert(ctx context.Context, ev domain.Event) error {
	meta := ev.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}

	var toolOrIntent any
	if ev.ToolOrIntent != "" {
		toolOrIntent = ev.ToolOrIntent
	}
	var model any
	if ev.Model != "" {
		model = ev.Model
	}
	var success any
	if ev.Success != nil {
		success = *ev.Success
	}

	const q = `
		INSERT INTO anon_learning_events (
			id, anon_subject, event_type, tool_or_intent, success,
			ask_rounds, model, meta, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9)
	`
	_, err = r.pool.Exec(ctx, q,
		ev.ID.UUID(),
		ev.AnonSubject,
		ev.Type,
		toolOrIntent,
		success,
		ev.AskRounds,
		model,
		metaJSON,
		ev.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert anon_learning_events: %w", err)
	}
	return nil
}
