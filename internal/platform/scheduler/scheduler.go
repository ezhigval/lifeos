package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
)

type JobHandler func(ctx context.Context, userID ids.UserID, payload json.RawMessage) error

type Scheduler struct {
	pool     *pgxpool.Pool
	log      *slog.Logger
	handlers map[string]JobHandler
	now      func() time.Time
	tick     time.Duration
}

func New(pool *pgxpool.Pool, log *slog.Logger) *Scheduler {
	return &Scheduler{
		pool:     pool,
		log:      log,
		handlers: make(map[string]JobHandler),
		now:      func() time.Time { return time.Now().UTC() },
		tick:     time.Minute,
	}
}

func (s *Scheduler) Register(jobType string, h JobHandler) {
	s.handlers[jobType] = h
}

func (s *Scheduler) Run(ctx context.Context) error {
	s.log.Info("scheduler started")
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.tickOnce(ctx); err != nil {
				s.log.Error("scheduler tick failed", "error", err)
			}
		}
	}
}

func (s *Scheduler) tickOnce(ctx context.Context) error {
	s.reclaimStuck(ctx)
	queries := db.New(s.pool)
	jobs, err := queries.ClaimDueJobs(ctx, db.ClaimDueJobsParams{
		RunAt: pgconv.TimestamptzValue(s.now()),
		Limit: 20,
	})
	if err != nil {
		return err
	}
	for _, job := range jobs {
		userID := pgconv.FromUserID(job.UserID)
		handler := s.handlers[job.JobType]
		if handler == nil {
			s.log.Error("unknown job type", "job_type", job.JobType)
			_ = queries.RescheduleJob(ctx, db.RescheduleJobParams{
				ID:    job.ID,
				RunAt: pgconv.TimestamptzValue(s.now().Add(time.Hour)),
			})
			continue
		}
		if err := handler(ctx, userID, job.Payload); err != nil {
			var deferErr DeferJobError
			if errors.As(err, &deferErr) {
				_ = queries.RescheduleJob(ctx, db.RescheduleJobParams{
					ID:    job.ID,
					RunAt: pgconv.TimestamptzValue(s.now().Add(deferErr.Delay)),
				})
				s.log.Info("job deferred", "job_type", job.JobType, "delay", deferErr.Delay)
				continue
			}
			s.log.Error("job failed", "job_type", job.JobType, "error", err)
			_ = queries.MarkJobPending(ctx, job.ID)
			continue
		}
		_ = queries.MarkJobDone(ctx, job.ID)
	}
	return nil
}

// reclaimStuck returns jobs left in processing after a crash so they can run again.
func (s *Scheduler) reclaimStuck(ctx context.Context) {
	_, err := s.pool.Exec(ctx, `
		UPDATE scheduled_jobs
		SET status = 'pending', updated_at = now()
		WHERE status = 'processing'
		  AND updated_at < now() - interval '15 minutes'`)
	if err != nil {
		s.log.Error("reclaim stuck jobs", "error", err)
	}
}
