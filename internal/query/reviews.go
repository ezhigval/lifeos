package query

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/ai"
	templateassistant "github.com/valentinezhov/lifeos/internal/ai/templateassistant"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	projectsinfra "github.com/valentinezhov/lifeos/internal/projects/infra"
	tasksinfra "github.com/valentinezhov/lifeos/internal/tasks/infra"
)

type Review struct {
	tasks     *tasksinfra.Repository
	projects  *projectsinfra.Repository
	assistant ai.Assistant
	users     UserTimezone
	period    *PeriodReview
	now       func() time.Time
}

func NewReview(pool *pgxpool.Pool, users UserTimezone, assistant ai.Assistant) *Review {
	now := func() time.Time { return time.Now().UTC() }
	if assistant == nil {
		assistant = templateassistant.New()
	}
	return &Review{
		tasks:    tasksinfra.NewRepository(pool),
		projects: projectsinfra.NewRepository(pool),
		assistant: assistant,
		users:    users,
		period: &PeriodReview{
			pool:     pool,
			projects: newProjectsReader(pool),
			users:    users,
			now:      now,
		},
		now: now,
	}
}

func (r *Review) Morning(ctx context.Context, userID ids.UserID) (string, error) {
	return r.build(ctx, userID, "🌅 <b>Утренний обзор</b>\n")
}

func (r *Review) Evening(ctx context.Context, userID ids.UserID) (string, error) {
	return r.build(ctx, userID, "🌙 <b>Вечерний обзор</b>\n")
}

func (r *Review) Weekly(ctx context.Context, userID ids.UserID) (string, error) {
	return r.period.Weekly(ctx, userID)
}

func (r *Review) Monthly(ctx context.Context, userID ids.UserID, previousMonth bool) (string, error) {
	return r.period.Monthly(ctx, userID, previousMonth)
}

func (r *Review) build(ctx context.Context, userID ids.UserID, header string) (string, error) {
	tz, err := r.users.Timezone(ctx, userID)
	if err != nil {
		return "", err
	}
	today, err := timeutil.DateInTimezone(r.now(), tz)
	if err != nil {
		return "", err
	}
	tasks, err := r.tasks.ListByDueDate(ctx, userID, today)
	if err != nil {
		return "", err
	}
	taskTitles := make([]string, 0, len(tasks))
	for _, t := range tasks {
		taskTitles = append(taskTitles, t.Title)
	}
	projects, err := r.projects.ListActive(ctx, userID)
	if err != nil {
		return "", err
	}
	projectTitles := make([]string, 0, len(projects))
	for _, p := range projects {
		projectTitles = append(projectTitles, p.Name)
	}
	body, err := r.assistant.Summarize(ctx, ai.SummaryRequest{Tasks: taskTitles, Projects: projectTitles})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", header, body), nil
}
