package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/calendar/domain"
	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, event domain.Event) error {
	endsAt := pgconv.Timestamptz(event.EndsAt)
	return r.queries(ctx).InsertCalendarEvent(ctx, db.InsertCalendarEventParams{
		ID:        pgconv.EventID(event.ID),
		UserID:    pgconv.UserID(event.UserID),
		Title:     event.Title,
		StartsAt:  pgconv.TimestamptzValue(event.StartsAt),
		EndsAt:    endsAt,
		CreatedAt: pgconv.TimestamptzValue(event.CreatedAt),
	})
}

func (r *Repository) ListBetween(ctx context.Context, userID ids.UserID, from, to time.Time) ([]domain.Event, error) {
	rows, err := r.queries(ctx).ListCalendarEventsBetween(ctx, db.ListCalendarEventsBetweenParams{
		UserID:     pgconv.UserID(userID),
		StartsAt:   pgconv.TimestamptzValue(from),
		StartsAt_2: pgconv.TimestamptzValue(to),
	})
	if err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}
	out := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapEvent(row))
	}
	return out, nil
}

func mapEvent(row db.CalendarEvent) domain.Event {
	var endsAt *time.Time
	if row.EndsAt.Valid {
		t := row.EndsAt.Time.UTC()
		endsAt = &t
	}
	return domain.Event{
		ID:        pgconv.FromEventID(row.ID),
		UserID:    pgconv.FromUserID(row.UserID),
		Title:     row.Title,
		StartsAt:  row.StartsAt.Time.UTC(),
		EndsAt:    endsAt,
		CreatedAt: row.CreatedAt.Time.UTC(),
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
