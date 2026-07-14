package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, task domain.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}
	tags := task.Tags
	if tags == nil {
		tags = []string{}
	}
	_, err := r.queries(ctx).InsertTask(ctx, db.InsertTaskParams{
		ID:              pgconv.TaskID(task.ID),
		UserID:          pgconv.UserID(task.UserID),
		Title:           task.Title,
		Description:     pgconv.Text(task.Description),
		Status:          string(task.Status),
		Priority:        string(task.Priority),
		DueDate:         pgconv.DatePtr(task.DueDate),
		DurationMinutes: pgconv.Int4Ptr(task.DurationMinutes),
		Tags:            tags,
		CreatedAt:       pgconv.TimestamptzValue(task.CreatedAt),
	})
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func (r *Repository) SetProjects(ctx context.Context, taskID ids.TaskID, projectIDs []ids.ProjectID) error {
	q := r.queries(ctx)
	if err := q.DeleteTaskProjects(ctx, pgconv.TaskID(taskID)); err != nil {
		return err
	}
	for _, pid := range projectIDs {
		if err := q.InsertTaskProject(ctx, db.InsertTaskProjectParams{
			TaskID:    pgconv.TaskID(taskID),
			ProjectID: pgconv.ProjectID(pid),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) loadProjectIDs(ctx context.Context, taskID ids.TaskID) ([]ids.ProjectID, error) {
	rows, err := r.queries(ctx).ListProjectIDsByTask(ctx, pgconv.TaskID(taskID))
	if err != nil {
		return nil, err
	}
	out := make([]ids.ProjectID, 0, len(rows))
	for _, row := range rows {
		out = append(out, pgconv.FromProjectID(row))
	}
	return out, nil
}

func (r *Repository) attachProjects(ctx context.Context, task domain.Task) (domain.Task, error) {
	ids, err := r.loadProjectIDs(ctx, task.ID)
	if err != nil {
		return domain.Task{}, err
	}
	task.ProjectIDs = ids
	return task, nil
}

func (r *Repository) GetByID(ctx context.Context, userID ids.UserID, taskID ids.TaskID) (domain.Task, error) {
	row, err := r.queries(ctx).GetTaskByID(ctx, db.GetTaskByIDParams{
		ID:     pgconv.TaskID(taskID),
		UserID: pgconv.UserID(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}
	return r.attachProjects(ctx, mapTask(row))
}

func (r *Repository) ListByDueDate(ctx context.Context, userID ids.UserID, dueDate time.Time) ([]domain.Task, error) {
	rows, err := r.queries(ctx).ListTasksByDueDate(ctx, db.ListTasksByDueDateParams{
		UserID:  pgconv.UserID(userID),
		DueDate: pgconv.Date(dueDate),
	})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return r.mapTasks(ctx, rows)
}

func (r *Repository) ListOpenDueOnOrBefore(ctx context.Context, userID ids.UserID, dueDate time.Time) ([]domain.Task, error) {
	rows, err := r.queries(ctx).ListOpenTasksDueOnOrBefore(ctx, db.ListOpenTasksDueOnOrBeforeParams{
		UserID:  pgconv.UserID(userID),
		DueDate: pgconv.Date(dueDate),
	})
	if err != nil {
		return nil, fmt.Errorf("list open tasks due on or before: %w", err)
	}
	return r.mapTasks(ctx, rows)
}

func (r *Repository) ListByTag(ctx context.Context, userID ids.UserID, tag string) ([]domain.Task, error) {
	rows, err := r.queries(ctx).ListTasksByTag(ctx, db.ListTasksByTagParams{
		UserID: pgconv.UserID(userID),
		Tag:    tag,
	})
	if err != nil {
		return nil, fmt.Errorf("list tasks by tag: %w", err)
	}
	return r.mapTasks(ctx, rows)
}

func (r *Repository) ListByProject(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) ([]domain.Task, error) {
	rows, err := r.queries(ctx).ListTasksByProjectJoin(ctx, db.ListTasksByProjectJoinParams{
		UserID:    pgconv.UserID(userID),
		ProjectID: pgconv.ProjectID(projectID),
	})
	if err != nil {
		return nil, fmt.Errorf("list tasks by project: %w", err)
	}
	return r.mapTasks(ctx, rows)
}

func (r *Repository) FindOpenByTitle(ctx context.Context, userID ids.UserID, title string) (domain.Task, error) {
	row, err := r.queries(ctx).FindOpenTaskByTitle(ctx, db.FindOpenTaskByTitleParams{
		UserID:  pgconv.UserID(userID),
		Column2: pgconv.Text(&title),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("find task: %w", err)
	}
	return r.attachProjects(ctx, mapTask(row))
}

func (r *Repository) Update(ctx context.Context, task domain.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}
	tags := task.Tags
	if tags == nil {
		tags = []string{}
	}
	_, err := r.queries(ctx).UpdateTask(ctx, db.UpdateTaskParams{
		ID:              pgconv.TaskID(task.ID),
		UserID:          pgconv.UserID(task.UserID),
		Title:           task.Title,
		Description:     pgconv.Text(task.Description),
		Status:          string(task.Status),
		Priority:        string(task.Priority),
		DueDate:         pgconv.DatePtr(task.DueDate),
		DurationMinutes: pgconv.Int4Ptr(task.DurationMinutes),
		Tags:            tags,
		CompletedAt:     pgconv.Timestamptz(task.CompletedAt),
		DeletedAt:       pgconv.Timestamptz(task.DeletedAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (r *Repository) mapTasks(ctx context.Context, rows []db.Task) ([]domain.Task, error) {
	out := make([]domain.Task, 0, len(rows))
	for _, row := range rows {
		task, err := r.attachProjects(ctx, mapTask(row))
		if err != nil {
			return nil, err
		}
		out = append(out, task)
	}
	return out, nil
}

func mapTask(row db.Task) domain.Task {
	tags := row.Tags
	if tags == nil {
		tags = []string{}
	}
	return domain.Task{
		ID:              pgconv.FromTaskID(row.ID),
		UserID:          pgconv.FromUserID(row.UserID),
		Title:           row.Title,
		Description:     pgconv.FromText(row.Description),
		Status:          domain.Status(row.Status),
		Priority:        domain.Priority(row.Priority),
		DueDate:         pgconv.FromDate(row.DueDate),
		DurationMinutes: pgconv.FromInt4Ptr(row.DurationMinutes),
		Tags:            tags,
		CompletedAt:     pgconv.FromTimestamptz(row.CompletedAt),
		DeletedAt:       pgconv.FromTimestamptz(row.DeletedAt),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
