package query

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	financedomain "github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
)

type PeriodStats struct {
	PeriodLabel      string
	CompletedTasks   int64
	OpenTasks        int64
	HabitCompletions int64
	IncomeCents      int64
	ExpenseCents     int64
	NetCents         int64
	ActiveProjects   int
}

type PeriodReview struct {
	pool     *pgxpool.Pool
	projects *projectsReader
	users    UserTimezone
	now      func() time.Time
}

type projectsReader struct {
	pool *pgxpool.Pool
}

func newProjectsReader(pool *pgxpool.Pool) *projectsReader {
	return &projectsReader{pool: pool}
}

func (r *projectsReader) countActive(ctx context.Context, userID ids.UserID) (int, error) {
	rows, err := db.New(r.pool).ListActiveProjects(ctx, pgconv.UserID(userID))
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (r *PeriodReview) stats(ctx context.Context, userID ids.UserID, start, end time.Time, tz, label string) (PeriodStats, error) {
	q := db.New(r.pool)
	uid := pgconv.UserID(userID)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	startDate := dateInLoc(start, loc)
	endDate := dateInLoc(end, loc)

	completed, err := q.CountCompletedTasksBetween(ctx, db.CountCompletedTasksBetweenParams{
		UserID:        uid,
		CompletedAt:   pgconv.TimestamptzValue(start),
		CompletedAt_2: pgconv.TimestamptzValue(end),
	})
	if err != nil {
		return PeriodStats{}, fmt.Errorf("count completed tasks: %w", err)
	}
	open, err := q.CountOpenTasks(ctx, uid)
	if err != nil {
		return PeriodStats{}, fmt.Errorf("count open tasks: %w", err)
	}
	habits, err := q.CountHabitCompletionsBetween(ctx, db.CountHabitCompletionsBetweenParams{
		UserID:    uid,
		LogDate:   pgconv.Date(startDate),
		LogDate_2: pgconv.Date(endDate),
	})
	if err != nil {
		return PeriodStats{}, fmt.Errorf("count habit completions: %w", err)
	}
	income, err := q.SumIncomeBetween(ctx, db.SumIncomeBetweenParams{
		UserID:       uid,
		OccurredAt:   pgconv.TimestamptzValue(start),
		OccurredAt_2: pgconv.TimestamptzValue(end),
	})
	if err != nil {
		return PeriodStats{}, fmt.Errorf("sum income: %w", err)
	}
	expense, err := q.SumExpenseBetween(ctx, db.SumExpenseBetweenParams{
		UserID:       uid,
		OccurredAt:   pgconv.TimestamptzValue(start),
		OccurredAt_2: pgconv.TimestamptzValue(end),
	})
	if err != nil {
		return PeriodStats{}, fmt.Errorf("sum expense: %w", err)
	}
	projects, err := r.projects.countActive(ctx, userID)
	if err != nil {
		return PeriodStats{}, fmt.Errorf("count projects: %w", err)
	}

	return PeriodStats{
		PeriodLabel:      label,
		CompletedTasks:   completed,
		OpenTasks:        open,
		HabitCompletions: habits,
		IncomeCents:      income,
		ExpenseCents:     expense,
		NetCents:         income - expense,
		ActiveProjects:   projects,
	}, nil
}

func (r *PeriodReview) Weekly(ctx context.Context, userID ids.UserID) (string, error) {
	tz, err := r.users.Timezone(ctx, userID)
	if err != nil {
		return "", err
	}
	start, end, err := timeutil.WeekBounds(r.now(), tz)
	if err != nil {
		return "", err
	}
	label := formatPeriodLabel(start, end, tz)
	stats, err := r.stats(ctx, userID, start, end, tz, label)
	if err != nil {
		return "", err
	}
	return formatPeriodReview("📊 <b>Недельный обзор</b>", stats), nil
}

func (r *PeriodReview) Monthly(ctx context.Context, userID ids.UserID, previousMonth bool) (string, error) {
	tz, err := r.users.Timezone(ctx, userID)
	if err != nil {
		return "", err
	}
	var start, end time.Time
	if previousMonth {
		start, end, err = timeutil.PreviousMonthBounds(r.now(), tz)
	} else {
		start, end, err = timeutil.MonthBounds(r.now(), tz)
	}
	if err != nil {
		return "", err
	}
	label := formatPeriodLabel(start, end, tz)
	stats, err := r.stats(ctx, userID, start, end, tz, label)
	if err != nil {
		return "", err
	}
	return formatPeriodReview("📈 <b>Месячный обзор</b>", stats), nil
}

func formatPeriodLabel(start, end time.Time, tz string) string {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	s := start.In(loc)
	e := end.Add(-time.Nanosecond).In(loc)
	return fmt.Sprintf("%s–%s", s.Format("02.01"), e.Format("02.01"))
}

func formatPeriodReview(header string, s PeriodStats) string {
	income := financedomain.FormatMoney(financedomain.Money{AmountCents: s.IncomeCents, Currency: "RUB"})
	expense := financedomain.FormatMoney(financedomain.Money{AmountCents: s.ExpenseCents, Currency: "RUB"})
	net := financedomain.FormatMoney(financedomain.Money{AmountCents: s.NetCents, Currency: "RUB"})
	return fmt.Sprintf(
		"%s (%s)\n✅ Задач закрыто: <b>%d</b>\n📋 Открытых задач: <b>%d</b>\n🔄 Отметок привычек: <b>%d</b>\n💰 Доход: %s | Расход: %s | Итого: %s\n📁 Активных проектов: <b>%d</b>",
		header, s.PeriodLabel,
		s.CompletedTasks, s.OpenTasks, s.HabitCompletions,
		income, expense, net,
		s.ActiveProjects,
	)
}

func dateInLoc(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}
