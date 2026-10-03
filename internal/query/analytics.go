package query

import (
	"context"
	"fmt"
	"html"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	projectsinfra "github.com/valentinezhov/lifeos/internal/projects/infra"
)

type ProjectKPI struct {
	Title   string `json:"title"`
	Percent string `json:"percent"`
}

type ProductivitySummary struct {
	PeriodLabel      string
	TasksCreated     int64
	TasksCompleted   int64
	CompletionRate   int
	OpenTasks        int64
	HabitConsistency int
	HabitCompletions int64
	HabitCount       int64
	Projects         []ProjectKPI
}

type GetProductivitySummary struct {
	pool     *pgxpool.Pool
	projects *projectsinfra.Repository
	users    UserTimezone
	now      func() time.Time
}

func NewGetProductivitySummary(pool *pgxpool.Pool, users UserTimezone) *GetProductivitySummary {
	return &GetProductivitySummary{
		pool:     pool,
		projects: projectsinfra.NewRepository(pool),
		users:    users,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (q *GetProductivitySummary) Execute(ctx context.Context, userID ids.UserID) (ProductivitySummary, error) {
	if userID.IsZero() {
		return ProductivitySummary{}, fmt.Errorf("user id is required")
	}
	tz, err := q.users.Timezone(ctx, userID)
	if err != nil {
		return ProductivitySummary{}, err
	}
	start, end, err := timeutil.MonthBounds(q.now(), tz)
	if err != nil {
		return ProductivitySummary{}, err
	}
	return q.forPeriod(ctx, userID, tz, start, end)
}

func (q *GetProductivitySummary) forPeriod(ctx context.Context, userID ids.UserID, tz string, start, end time.Time) (ProductivitySummary, error) {
	dbq := db.New(q.pool)
	uid := pgconv.UserID(userID)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	startDate := dateInLoc(start, loc)
	endDate := dateInLoc(end, loc)

	created, err := dbq.CountTasksCreatedBetween(ctx, db.CountTasksCreatedBetweenParams{
		UserID:      uid,
		CreatedAt:   pgconv.TimestamptzValue(start),
		CreatedAt_2: pgconv.TimestamptzValue(end),
	})
	if err != nil {
		return ProductivitySummary{}, fmt.Errorf("count created tasks: %w", err)
	}
	completed, err := dbq.CountCompletedTasksBetween(ctx, db.CountCompletedTasksBetweenParams{
		UserID:        uid,
		CompletedAt:   pgconv.TimestamptzValue(start),
		CompletedAt_2: pgconv.TimestamptzValue(end),
	})
	if err != nil {
		return ProductivitySummary{}, fmt.Errorf("count completed tasks: %w", err)
	}
	open, err := dbq.CountOpenTasks(ctx, uid)
	if err != nil {
		return ProductivitySummary{}, fmt.Errorf("count open tasks: %w", err)
	}
	habitCount, err := dbq.CountUserHabits(ctx, uid)
	if err != nil {
		return ProductivitySummary{}, fmt.Errorf("count habits: %w", err)
	}
	habitDone, err := dbq.CountHabitCompletionsBetween(ctx, db.CountHabitCompletionsBetweenParams{
		UserID:    uid,
		LogDate:   pgconv.Date(startDate),
		LogDate_2: pgconv.Date(endDate),
	})
	if err != nil {
		return ProductivitySummary{}, fmt.Errorf("count habit completions: %w", err)
	}

	daysInPeriod := daysBetween(startDate, dateInLoc(q.now(), loc))
	if daysInPeriod < 1 {
		daysInPeriod = 1
	}
	habitConsistency := 0
	if habitCount > 0 {
		expected := habitCount * int64(daysInPeriod)
		habitConsistency = int(math.Round(float64(habitDone) * 100 / float64(expected)))
		if habitConsistency > 100 {
			habitConsistency = 100
		}
	}

	completionRate := 0
	if created > 0 {
		completionRate = int(math.Round(float64(completed) * 100 / float64(created)))
	} else if completed > 0 {
		completionRate = 100
	}

	kpis, err := q.projectKPIs(ctx, userID)
	if err != nil {
		return ProductivitySummary{}, err
	}

	return ProductivitySummary{
		PeriodLabel:      formatPeriodLabel(start, end, tz),
		TasksCreated:     created,
		TasksCompleted:   completed,
		CompletionRate:   completionRate,
		OpenTasks:        open,
		HabitConsistency: habitConsistency,
		HabitCompletions: habitDone,
		HabitCount:       habitCount,
		Projects:         kpis,
	}, nil
}

func (q *GetProductivitySummary) projectKPIs(ctx context.Context, userID ids.UserID) ([]ProjectKPI, error) {
	projects, err := q.projects.ListActive(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	out := make([]ProjectKPI, 0, len(projects))
	for _, project := range projects {
		item := ProjectKPI{Title: project.Name}
		progress, err := project.Progress()
		if err == nil && progress.HasTarget {
			item.Percent = progress.Percent.StringFixed(0) + "%"
		} else {
			item.Percent = "—"
		}
		out = append(out, item)
		if len(out) >= 5 {
			break
		}
	}
	return out, nil
}

func daysBetween(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func FormatProductivitySummary(s ProductivitySummary) string {
	var projectsBlock string
	if len(s.Projects) == 0 {
		projectsBlock = "Проектов нет."
	} else {
		for _, p := range s.Projects {
			projectsBlock += fmt.Sprintf("• %s — %s\n", html.EscapeString(p.Title), p.Percent)
		}
	}
	return fmt.Sprintf(
		"📊 <b>Статистика</b> (%s)\n\n"+
			"✅ <b>Задачи</b>\nСоздано: %d | Закрыто: %d | Rate: <b>%d%%</b>\nОткрытых сейчас: %d\n\n"+
			"🔄 <b>Привычки</b>\nОтметок: %d | Consistency: <b>%d%%</b> (%d привычек)\n\n"+
			"📁 <b>Проекты</b>\n%s",
		s.PeriodLabel,
		s.TasksCreated, s.TasksCompleted, s.CompletionRate, s.OpenTasks,
		s.HabitCompletions, s.HabitConsistency, s.HabitCount,
		projectsBlock,
	)
}
