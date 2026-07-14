package infra

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/settings/domain"
)

type QuietHours struct {
	repo *Repository
}

func NewQuietHours(pool *pgxpool.Pool) *QuietHours {
	return &QuietHours{repo: NewRepository(pool)}
}

func (q *QuietHours) InQuietPeriod(ctx context.Context, userID ids.UserID, now time.Time) (bool, error) {
	s, err := q.repo.Get(ctx, userID)
	if err != nil || s.QuietHoursStart == nil || s.QuietHoursEnd == nil {
		return false, err
	}
	cur := domain.TimeOfDay{Hour: now.Hour(), Minute: now.Minute()}
	return inQuietRange(cur, *s.QuietHoursStart, *s.QuietHoursEnd), nil
}

func inQuietRange(cur, start, end domain.TimeOfDay) bool {
	c := cur.Hour*60 + cur.Minute
	s := start.Hour*60 + start.Minute
	e := end.Hour*60 + end.Minute
	if s < e {
		return c >= s && c < e
	}
	return c >= s || c < e
}
