package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/habits/app"
	"github.com/valentinezhov/lifeos/internal/habits/domain"
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

func (r *Repository) Save(ctx context.Context, habit domain.Habit) error {
	return r.queries(ctx).InsertHabit(ctx, db.InsertHabitParams{
		ID:        pgconv.HabitID(habit.ID),
		UserID:    pgconv.UserID(habit.UserID),
		Name:      habit.Name,
		Frequency: string(habit.Frequency),
		StartDate: pgconv.DatePtr(habit.StartDate),
		EndDate:   pgconv.DatePtr(habit.EndDate),
		CreatedAt: pgconv.TimestamptzValue(habit.CreatedAt),
	})
}

// Update overwrites editable fields; returns false when the habit does not exist
// for this user.
func (r *Repository) Update(ctx context.Context, habit domain.Habit) (bool, error) {
	rows, err := r.queries(ctx).UpdateHabit(ctx, db.UpdateHabitParams{
		ID:        pgconv.HabitID(habit.ID),
		UserID:    pgconv.UserID(habit.UserID),
		Name:      habit.Name,
		Frequency: string(habit.Frequency),
		StartDate: pgconv.DatePtr(habit.StartDate),
		EndDate:   pgconv.DatePtr(habit.EndDate),
	})
	if err != nil {
		return false, fmt.Errorf("update habit: %w", err)
	}
	return rows > 0, nil
}

// Delete removes the habit (logs cascade via FK); returns false when not found.
func (r *Repository) Delete(ctx context.Context, userID ids.UserID, habitID ids.HabitID) (bool, error) {
	rows, err := r.queries(ctx).DeleteHabit(ctx, db.DeleteHabitParams{
		ID:     pgconv.HabitID(habitID),
		UserID: pgconv.UserID(userID),
	})
	if err != nil {
		return false, fmt.Errorf("delete habit: %w", err)
	}
	return rows > 0, nil
}

func (r *Repository) GetByID(ctx context.Context, userID ids.UserID, habitID ids.HabitID) (domain.Habit, error) {
	row, err := r.queries(ctx).GetHabitByID(ctx, db.GetHabitByIDParams{
		ID:     pgconv.HabitID(habitID),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Habit{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Habit{}, fmt.Errorf("get habit: %w", err)
	}
	return mapHabit(row), nil
}

func (r *Repository) FindByName(ctx context.Context, userID ids.UserID, name string) (domain.Habit, error) {
	row, err := r.queries(ctx).FindHabitByName(ctx, db.FindHabitByNameParams{
		UserID: pgconv.UserID(userID),
		Lower:  name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Habit{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Habit{}, fmt.Errorf("find habit: %w", err)
	}
	return mapHabit(row), nil
}

func (r *Repository) ListWithToday(ctx context.Context, userID ids.UserID, today time.Time) ([]app.HabitDayRow, error) {
	rows, err := r.queries(ctx).ListHabitsWithTodayLog(ctx, db.ListHabitsWithTodayLogParams{
		UserID:  pgconv.UserID(userID),
		LogDate: pgconv.Date(today),
	})
	if err != nil {
		return nil, fmt.Errorf("list habits today: %w", err)
	}
	out := make([]app.HabitDayRow, 0, len(rows))
	for _, row := range rows {
		completed := false
		if row.TodayCompleted.Valid {
			completed = row.TodayCompleted.Bool
		}
		out = append(out, app.HabitDayRow{
			Habit: domain.Habit{
				ID:        pgconv.FromHabitID(row.ID),
				UserID:    pgconv.FromUserID(row.UserID),
				Name:      row.Name,
				Frequency: domain.Frequency(row.Frequency),
				StartDate: pgconv.FromDate(row.StartDate),
				EndDate:   pgconv.FromDate(row.EndDate),
				CreatedAt: row.CreatedAt.Time,
			},
			TodayCompleted: completed,
		})
	}
	return out, nil
}

func (r *Repository) Upsert(ctx context.Context, log domain.HabitLog) error {
	return r.queries(ctx).UpsertHabitLog(ctx, db.UpsertHabitLogParams{
		ID:        pgconv.UUID(log.ID.UUID()),
		HabitID:   pgconv.HabitID(log.HabitID),
		LogDate:   pgconv.Date(log.LogDate),
		Completed: log.Completed,
		CreatedAt: pgconv.TimestamptzValue(log.CreatedAt),
	})
}

func (r *Repository) ListSince(ctx context.Context, habitID ids.HabitID, since time.Time) ([]domain.HabitLog, error) {
	rows, err := r.queries(ctx).ListHabitLogsSince(ctx, db.ListHabitLogsSinceParams{
		HabitID: pgconv.HabitID(habitID),
		LogDate: pgconv.Date(since),
	})
	if err != nil {
		return nil, fmt.Errorf("list habit logs: %w", err)
	}
	out := make([]domain.HabitLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.HabitLog{
			ID:        ids.HabitLogID(pgconv.FromUUID(row.ID)),
			HabitID:   pgconv.FromHabitID(row.HabitID),
			LogDate:   row.LogDate.Time,
			Completed: row.Completed,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return out, nil
}

func mapHabit(row db.Habit) domain.Habit {
	return domain.Habit{
		ID:        pgconv.FromHabitID(row.ID),
		UserID:    pgconv.FromUserID(row.UserID),
		Name:      row.Name,
		Frequency: domain.Frequency(row.Frequency),
		StartDate: pgconv.FromDate(row.StartDate),
		EndDate:   pgconv.FromDate(row.EndDate),
		CreatedAt: row.CreatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
