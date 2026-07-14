package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/planning/domain"
	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Upsert(ctx context.Context, a domain.DayAvailability) error {
	until := pgtype.Time{Valid: true, Microseconds: int64(a.AvailableUntil.Hour()*3600+a.AvailableUntil.Minute()*60) * 1e6}
	_, err := db.New(r.pool).UpsertDayAvailability(ctx, db.UpsertDayAvailabilityParams{
		ID:             pgconv.TaskID(a.ID),
		UserID:         pgconv.UserID(a.UserID),
		Day:            pgconv.Date(a.Day),
		AvailableUntil: until,
		Note:           pgconv.Text(a.Note),
		CreatedAt:      pgconv.TimestamptzValue(time.Now().UTC()),
	})
	if err != nil {
		return fmt.Errorf("upsert availability: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, userID ids.UserID, day time.Time) (domain.DayAvailability, error) {
	row, err := db.New(r.pool).GetDayAvailability(ctx, db.GetDayAvailabilityParams{
		UserID: pgconv.UserID(userID),
		Day:    pgconv.Date(day),
	})
	if err != nil {
		return domain.DayAvailability{}, err
	}
	until := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	if row.AvailableUntil.Valid {
		sec := row.AvailableUntil.Microseconds / 1e6
		until = until.Add(time.Duration(sec) * time.Second)
	}
	return domain.DayAvailability{
		ID:             pgconv.FromTaskID(row.ID),
		UserID:         pgconv.FromUserID(row.UserID),
		Day:            pgconv.FromDate(row.Day).UTC(),
		AvailableUntil: until,
		Note:           pgconv.FromText(row.Note),
	}, nil
}
