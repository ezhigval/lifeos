package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
)

var ErrReminderNotFound = errors.New("reminder not found")

type CancelReminder struct {
	pool *pgxpool.Pool
}

func NewCancelReminder(pool *pgxpool.Pool) *CancelReminder {
	return &CancelReminder{pool: pool}
}

type CancelReminderInput struct {
	UserID     ids.UserID
	ReminderID uuid.UUID
}

func (uc *CancelReminder) Execute(ctx context.Context, in CancelReminderInput) (ReminderDTO, error) {
	if in.UserID.IsZero() || in.ReminderID == uuid.Nil {
		return ReminderDTO{}, fmt.Errorf("invalid cancel reminder input")
	}
	row, err := db.New(uc.pool).CancelPendingReminder(ctx, db.CancelPendingReminderParams{
		ID:     pgconv.UUID(in.ReminderID),
		UserID: pgconv.UserID(in.UserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReminderDTO{}, ErrReminderNotFound
		}
		return ReminderDTO{}, fmt.Errorf("cancel reminder: %w", err)
	}
	return reminderRowToDTO(row.ID, row.Payload, row.RunAt, row.Status), nil
}

// CancelForTask cancels pending reminder jobs linked to a task via payload.task_id.
func (uc *CancelReminder) CancelForTask(ctx context.Context, userID ids.UserID, taskID string) error {
	if userID.IsZero() || strings.TrimSpace(taskID) == "" {
		return fmt.Errorf("invalid cancel-for-task input")
	}
	_, err := uc.pool.Exec(ctx, `
		UPDATE scheduled_jobs
		SET status = 'cancelled', updated_at = now()
		WHERE user_id = $1
		  AND job_type = 'reminder'
		  AND status = 'pending'
		  AND payload->>'task_id' = $2
	`, pgconv.UserID(userID), strings.TrimSpace(taskID))
	if err != nil {
		return fmt.Errorf("cancel reminders for task: %w", err)
	}
	return nil
}

func reminderRowToDTO(id pgtype.UUID, payload []byte, runAt pgtype.Timestamptz, status string) ReminderDTO {
	var p struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(payload, &p)
	return ReminderDTO{
		ID:      pgconv.FromUUID(id).String(),
		Message: p.Message,
		FireAt:  runAt.Time,
		Status:  status,
	}
}
