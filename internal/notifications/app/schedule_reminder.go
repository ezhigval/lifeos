package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
)

type JobStore struct {
	pool *pgxpool.Pool
}

func NewJobStore(pool *pgxpool.Pool) *JobStore {
	return &JobStore{pool: pool}
}

type ScheduleReminder struct {
	store *JobStore
	now   func() time.Time
}

func NewScheduleReminder(store *JobStore) *ScheduleReminder {
	return &ScheduleReminder{store: store, now: func() time.Time { return time.Now().UTC() }}
}

type ScheduleReminderInput struct {
	UserID  ids.UserID
	Message string
	FireAt  time.Time
}

func (uc *ScheduleReminder) Execute(ctx context.Context, in ScheduleReminderInput) error {
	if in.UserID.IsZero() || in.Message == "" {
		return fmt.Errorf("invalid reminder input")
	}
	payload, _ := json.Marshal(map[string]string{"message": in.Message})
	_, err := db.New(uc.store.pool).InsertScheduledJob(ctx, db.InsertScheduledJobParams{
		ID:      pgconv.UUID(uuid.Must(uuid.NewV7())),
		UserID:  pgconv.UserID(in.UserID),
		JobType: "reminder",
		Payload: payload,
		RunAt:   pgconv.TimestamptzValue(in.FireAt),
		Channel: "telegram",
	})
	if err != nil {
		return fmt.Errorf("schedule reminder: %w", err)
	}
	return nil
}

func (uc *ScheduleReminder) EnsureReview(ctx context.Context, userID ids.UserID, jobType string, runAt time.Time) error {
	q := db.New(uc.store.pool)
	ok, err := q.HasPendingJob(ctx, db.HasPendingJobParams{
		UserID:  pgconv.UserID(userID),
		JobType: jobType,
	})
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	payload, _ := json.Marshal(map[string]string{"type": jobType})
	_, err = q.InsertScheduledJob(ctx, db.InsertScheduledJobParams{
		ID:      pgconv.UUID(uuid.Must(uuid.NewV7())),
		UserID:  pgconv.UserID(userID),
		JobType: jobType,
		Payload: payload,
		RunAt:   pgconv.TimestamptzValue(runAt),
		Channel: "telegram",
	})
	return err
}

func (uc *ScheduleReminder) ScheduleNextReview(ctx context.Context, userID ids.UserID, jobType string, runAt time.Time) error {
	return uc.EnsureReview(ctx, userID, jobType, runAt)
}
