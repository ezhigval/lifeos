package app

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
)

type ReviewScheduler struct {
	pool *pgxpool.Pool
}

func NewReviewScheduler(pool *pgxpool.Pool) *ReviewScheduler {
	return &ReviewScheduler{pool: pool}
}

func (s *ReviewScheduler) RescheduleReview(ctx context.Context, userID ids.UserID, jobType string, runAt time.Time) error {
	return db.New(s.pool).ReschedulePendingJobsByType(ctx, db.ReschedulePendingJobsByTypeParams{
		UserID:  pgconv.UserID(userID),
		JobType: jobType,
		RunAt:   pgconv.TimestamptzValue(runAt),
	})
}
