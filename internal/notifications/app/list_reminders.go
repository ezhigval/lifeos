package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
)

type ReminderDTO struct {
	ID      string
	Message string
	FireAt  time.Time
	Status  string
}

type ListReminders struct {
	pool *pgxpool.Pool
}

func NewListReminders(pool *pgxpool.Pool) *ListReminders {
	return &ListReminders{pool: pool}
}

func (uc *ListReminders) Execute(ctx context.Context, userID ids.UserID) ([]ReminderDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	rows, err := db.New(uc.pool).ListPendingRemindersByUser(ctx, pgconv.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("list reminders: %w", err)
	}
	out := make([]ReminderDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, reminderRowToDTO(row.ID, row.Payload, row.RunAt, row.Status))
	}
	return out, nil
}
