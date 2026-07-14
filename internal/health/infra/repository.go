package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/health/domain"
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

func (r *Repository) Save(ctx context.Context, log domain.WeightLog) error {
	return r.queries(ctx).InsertWeightLog(ctx, db.InsertWeightLogParams{
		ID:        pgconv.WeightLogID(log.ID),
		UserID:    pgconv.UserID(log.UserID),
		WeightKg:  log.WeightKg,
		LoggedAt:  pgconv.TimestamptzValue(log.LoggedAt),
		CreatedAt: pgconv.TimestamptzValue(log.CreatedAt),
	})
}

func (r *Repository) GetLatest(ctx context.Context, userID ids.UserID) (domain.WeightLog, error) {
	row, err := r.queries(ctx).GetLatestWeightByUser(ctx, pgconv.UserID(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WeightLog{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.WeightLog{}, fmt.Errorf("get latest weight: %w", err)
	}
	return mapRow(row), nil
}

func (r *Repository) ListRecent(ctx context.Context, userID ids.UserID, limit int32) ([]domain.WeightLog, error) {
	rows, err := r.queries(ctx).ListRecentWeightsByUser(ctx, db.ListRecentWeightsByUserParams{
		UserID: pgconv.UserID(userID),
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list weights: %w", err)
	}
	out := make([]domain.WeightLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}
	return out, nil
}

func mapRow(row db.HealthWeightLog) domain.WeightLog {
	return domain.WeightLog{
		ID:        pgconv.FromWeightLogID(row.ID),
		UserID:    pgconv.FromUserID(row.UserID),
		WeightKg:  row.WeightKg,
		LoggedAt:  row.LoggedAt.Time,
		CreatedAt: row.CreatedAt.Time,
	}
}

func (r *Repository) SaveSteps(ctx context.Context, log domain.StepLog) error {
	return r.queries(ctx).InsertStepLog(ctx, db.InsertStepLogParams{
		ID:        pgconv.StepLogID(log.ID),
		UserID:    pgconv.UserID(log.UserID),
		Steps:     log.Steps,
		LoggedAt:  pgconv.TimestamptzValue(log.LoggedAt),
		CreatedAt: pgconv.TimestamptzValue(log.CreatedAt),
	})
}

func (r *Repository) GetLatestSteps(ctx context.Context, userID ids.UserID) (domain.StepLog, error) {
	row, err := r.queries(ctx).GetLatestStepsByUser(ctx, pgconv.UserID(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.StepLog{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.StepLog{}, fmt.Errorf("get latest steps: %w", err)
	}
	return mapStepRow(row), nil
}

func (r *Repository) ListRecentSteps(ctx context.Context, userID ids.UserID, limit int32) ([]domain.StepLog, error) {
	rows, err := r.queries(ctx).ListRecentStepsByUser(ctx, db.ListRecentStepsByUserParams{
		UserID: pgconv.UserID(userID),
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list steps: %w", err)
	}
	out := make([]domain.StepLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapStepRow(row))
	}
	return out, nil
}

func mapStepRow(row db.HealthStepLog) domain.StepLog {
	return domain.StepLog{
		ID:        pgconv.FromStepLogID(row.ID),
		UserID:    pgconv.FromUserID(row.UserID),
		Steps:     row.Steps,
		LoggedAt:  row.LoggedAt.Time,
		CreatedAt: row.CreatedAt.Time,
	}
}

func (r *Repository) SaveSleep(ctx context.Context, log domain.SleepLog) error {
	return r.queries(ctx).InsertSleepLog(ctx, db.InsertSleepLogParams{
		ID:              pgconv.SleepLogID(log.ID),
		UserID:          pgconv.UserID(log.UserID),
		DurationMinutes: log.DurationMinutes,
		LoggedAt:        pgconv.TimestamptzValue(log.LoggedAt),
		CreatedAt:       pgconv.TimestamptzValue(log.CreatedAt),
	})
}

func (r *Repository) GetLatestSleep(ctx context.Context, userID ids.UserID) (domain.SleepLog, error) {
	row, err := r.queries(ctx).GetLatestSleepByUser(ctx, pgconv.UserID(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SleepLog{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SleepLog{}, fmt.Errorf("get latest sleep: %w", err)
	}
	return mapSleepRow(row), nil
}

func (r *Repository) ListRecentSleep(ctx context.Context, userID ids.UserID, limit int32) ([]domain.SleepLog, error) {
	rows, err := r.queries(ctx).ListRecentSleepByUser(ctx, db.ListRecentSleepByUserParams{
		UserID: pgconv.UserID(userID),
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list sleep: %w", err)
	}
	out := make([]domain.SleepLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSleepRow(row))
	}
	return out, nil
}

func mapSleepRow(row db.HealthSleepLog) domain.SleepLog {
	return domain.SleepLog{
		ID:              pgconv.FromSleepLogID(row.ID),
		UserID:          pgconv.FromUserID(row.UserID),
		DurationMinutes: row.DurationMinutes,
		LoggedAt:        row.LoggedAt.Time,
		CreatedAt:       row.CreatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
