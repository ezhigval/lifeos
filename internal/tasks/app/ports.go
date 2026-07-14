package app

import (
	"context"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type TaskStore interface {
	Save(ctx context.Context, task domain.Task) error
	SetProjects(ctx context.Context, taskID ids.TaskID, projectIDs []ids.ProjectID) error
	GetByID(ctx context.Context, userID ids.UserID, taskID ids.TaskID) (domain.Task, error)
	ListByDueDate(ctx context.Context, userID ids.UserID, dueDate time.Time) ([]domain.Task, error)
	ListOpenDueOnOrBefore(ctx context.Context, userID ids.UserID, dueDate time.Time) ([]domain.Task, error)
	ListByTag(ctx context.Context, userID ids.UserID, tag string) ([]domain.Task, error)
	ListByProject(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) ([]domain.Task, error)
	FindOpenByTitle(ctx context.Context, userID ids.UserID, title string) (domain.Task, error)
	Update(ctx context.Context, task domain.Task) error
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type TaskDTO struct {
	ID              ids.TaskID
	Title           string
	Status          domain.Status
	Priority        domain.Priority
	DueDate         *time.Time
	DurationMinutes *int
	Tags            []string
	ProjectIDs      []ids.ProjectID
	CreatedAt       time.Time
}

func ToDTO(task domain.Task) TaskDTO {
	idsCopy := append([]ids.ProjectID(nil), task.ProjectIDs...)
	tagsCopy := append([]string(nil), task.Tags...)
	return TaskDTO{
		ID: task.ID, Title: task.Title, Status: task.Status, Priority: task.Priority,
		DueDate: task.DueDate, DurationMinutes: task.DurationMinutes, Tags: tagsCopy,
		ProjectIDs: idsCopy, CreatedAt: task.CreatedAt,
	}
}

func ToDTOs(tasks []domain.Task) []TaskDTO {
	out := make([]TaskDTO, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, ToDTO(task))
	}
	return out
}

type ProjectChecker interface {
	AllExist(ctx context.Context, userID ids.UserID, projectIDs []ids.ProjectID) (bool, error)
}
