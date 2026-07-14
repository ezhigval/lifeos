package query

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	projectsdomain "github.com/valentinezhov/lifeos/internal/projects/domain"
	projectsinfra "github.com/valentinezhov/lifeos/internal/projects/infra"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type PriorityItem struct {
	Kind   string
	Title  string
	Score  int
	Detail string
}

type UserTimezone interface {
	Timezone(ctx context.Context, userID ids.UserID) (string, error)
}

type GetTopPriorities struct {
	pool     *pgxpool.Pool
	projects *projectsinfra.Repository
	users    UserTimezone
	now      func() time.Time
}

func NewGetTopPriorities(pool *pgxpool.Pool, users UserTimezone) *GetTopPriorities {
	return &GetTopPriorities{
		pool:     pool,
		projects: projectsinfra.NewRepository(pool),
		users:    users,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (q *GetTopPriorities) Execute(ctx context.Context, userID ids.UserID) ([]PriorityItem, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}

	tz, err := q.users.Timezone(ctx, userID)
	if err != nil {
		return nil, err
	}
	today, err := timeutil.DateInTimezone(q.now(), tz)
	if err != nil {
		return nil, err
	}

	queries := db.New(q.pool)
	taskRows, err := queries.ListOverdueAndTodayTasks(ctx, db.ListOverdueAndTodayTasksParams{
		UserID:  pgconv.UserID(userID),
		DueDate: pgconv.Date(today),
	})
	if err != nil {
		return nil, fmt.Errorf("list priority tasks: %w", err)
	}

	items := make([]PriorityItem, 0, len(taskRows)+3)
	for _, row := range taskRows {
		score := priorityScore(taskdomain.Priority(row.Priority), pgconv.FromDate(row.DueDate), today)
		items = append(items, PriorityItem{
			Kind:   "task",
			Title:  row.Title,
			Score:  score,
			Detail: row.Priority,
		})
	}

	projects, err := q.projects.ListActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		score := projectScore(p)
		if score > 0 {
			items = append(items, PriorityItem{
				Kind:   "project",
				Title:  p.Name,
				Score:  score,
				Detail: string(p.Status),
			})
		}
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Score > items[j].Score })
	if len(items) > 5 {
		items = items[:5]
	}
	return items, nil
}

func priorityScore(p taskdomain.Priority, due *time.Time, today time.Time) int {
	score := 10
	switch p {
	case taskdomain.PriorityUrgent:
		score += 40
	case taskdomain.PriorityHigh:
		score += 25
	case taskdomain.PriorityMedium:
		score += 10
	}
	if due != nil {
		if due.Before(today) {
			score += 30
		} else if due.Equal(today) {
			score += 20
		}
	}
	return score
}

func projectScore(p projectsdomain.Project) int {
	if p.TargetValue == nil || p.TargetValue.IsZero() {
		return 5
	}
	progress, err := p.Progress()
	if err != nil {
		return 0
	}
	pct, _ := progress.Percent.Float64()
	if pct >= 80 {
		return 8
	}
	return int(20 - pct/5)
}
