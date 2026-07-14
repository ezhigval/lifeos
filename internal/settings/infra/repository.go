package infra

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	"github.com/valentinezhov/lifeos/internal/settings/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) EnsureDefaults(ctx context.Context, userID ids.UserID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_settings (user_id) VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, userID.UUID())
	return err
}

func (r *Repository) Get(ctx context.Context, userID ids.UserID) (domain.UserSettings, error) {
	row, err := db.New(r.pool).GetUserSettingsByUserID(ctx, pgconv.UserID(userID))
	if err != nil {
		return domain.UserSettings{}, err
	}
	return domain.UserSettings{
		UserID:          pgconv.FromUserID(row.UserID),
		MorningReviewAt: pgTimeToDomain(row.MorningReviewAt),
		EveningReviewAt: pgTimeToDomain(row.EveningReviewAt),
		WeeklyReviewAt:  pgTimeToDomain(row.WeeklyReviewAt),
		MonthlyReviewAt: pgTimeToDomain(row.MonthlyReviewAt),
		QuietHoursStart: pgTimePtr(row.QuietHoursStart),
		QuietHoursEnd:   pgTimePtr(row.QuietHoursEnd),
		Language:        row.Language,
	}, nil
}

func pgTimeToDomain(t pgtype.Time) domain.TimeOfDay {
	if !t.Valid {
		return domain.TimeOfDay{Hour: 8, Minute: 0}
	}
	sec := t.Microseconds / 1_000_000
	return domain.TimeOfDay{Hour: int(sec / 3600), Minute: int((sec % 3600) / 60)}
}

func pgTimePtr(t pgtype.Time) *domain.TimeOfDay {
	if !t.Valid {
		return nil
	}
	v := pgTimeToDomain(t)
	return &v
}

func (r *Repository) UpdateMorningReview(ctx context.Context, userID ids.UserID, at domain.TimeOfDay) error {
	return db.New(r.pool).UpdateMorningReviewAt(ctx, db.UpdateMorningReviewAtParams{
		UserID:          pgconv.UserID(userID),
		MorningReviewAt: domainToPGTime(at),
	})
}

func (r *Repository) UpdateEveningReview(ctx context.Context, userID ids.UserID, at domain.TimeOfDay) error {
	return db.New(r.pool).UpdateEveningReviewAt(ctx, db.UpdateEveningReviewAtParams{
		UserID:          pgconv.UserID(userID),
		EveningReviewAt: domainToPGTime(at),
	})
}

func (r *Repository) UpdateQuietHours(ctx context.Context, userID ids.UserID, start, end domain.TimeOfDay) error {
	return db.New(r.pool).UpdateQuietHours(ctx, db.UpdateQuietHoursParams{
		UserID:          pgconv.UserID(userID),
		QuietHoursStart: domainToPGTime(start),
		QuietHoursEnd:   domainToPGTime(end),
	})
}

func domainToPGTime(tod domain.TimeOfDay) pgtype.Time {
	var t pgtype.Time
	_ = t.Scan(domain.ToPGTime(tod))
	return t
}

// ReviewAt combines user settings time with a reference day in user timezone.
func ReviewAt(ref time.Time, tz string, tod domain.TimeOfDay) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	local := ref.In(loc)
	at := time.Date(local.Year(), local.Month(), local.Day(), tod.Hour, tod.Minute, 0, 0, loc)
	if !at.After(ref) {
		at = at.Add(24 * time.Hour)
	}
	return at.UTC(), nil
}

// NextWeeklyReviewAt returns the next Sunday at the configured time.
func NextWeeklyReviewAt(ref time.Time, tz string, tod domain.TimeOfDay) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	local := ref.In(loc)
	daysUntilSunday := (7 - int(local.Weekday())) % 7
	if daysUntilSunday == 0 {
		candidate := time.Date(local.Year(), local.Month(), local.Day(), tod.Hour, tod.Minute, 0, 0, loc)
		if !candidate.After(ref) {
			daysUntilSunday = 7
		}
	}
	sunday := local.AddDate(0, 0, daysUntilSunday)
	at := time.Date(sunday.Year(), sunday.Month(), sunday.Day(), tod.Hour, tod.Minute, 0, 0, loc)
	return at.UTC(), nil
}

// NextMonthlyReviewAt returns the next 1st day of month at the configured time.
func NextMonthlyReviewAt(ref time.Time, tz string, tod domain.TimeOfDay) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	local := ref.In(loc)
	candidate := time.Date(local.Year(), local.Month(), 1, tod.Hour, tod.Minute, 0, 0, loc)
	if !candidate.After(ref) {
		nextMonth := local.AddDate(0, 1, 0)
		candidate = time.Date(nextMonth.Year(), nextMonth.Month(), 1, tod.Hour, tod.Minute, 0, 0, loc)
	}
	return candidate.UTC(), nil
}
