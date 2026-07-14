package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyTitle        = errors.New("task title is required")
	ErrCannotComplete    = errors.New("cancelled task cannot be completed")
	ErrAlreadyDone       = errors.New("task is already done")
	ErrInvalidStatus     = errors.New("invalid task status")
	ErrInvalidPriority   = errors.New("invalid task priority")
	ErrCompletedAtNeeded = errors.New("completed task requires completed_at")
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type Task struct {
	ID          ids.TaskID
	UserID      ids.UserID
	Title       string
	Description *string
	Status      Status
	Priority    Priority
	DueDate     *time.Time
	ProjectIDs  []ids.ProjectID
	CompletedAt *time.Time
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTask(userID ids.UserID, title string, priority Priority, dueDate *time.Time, now time.Time) (Task, error) {
	if title == "" {
		return Task{}, ErrEmptyTitle
	}
	if !priority.Valid() {
		return Task{}, ErrInvalidPriority
	}

	return Task{
		ID:         ids.NewTaskID(),
		UserID:     userID,
		Title:      title,
		Status:     StatusTodo,
		Priority:   priority,
		DueDate:    dueDate,
		ProjectIDs: []ids.ProjectID{},
		CreatedAt:  now.UTC(),
		UpdatedAt:  now.UTC(),
	}, nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone, StatusCancelled:
		return true
	default:
		return false
	}
}

func (p Priority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

func (t *Task) Complete(now time.Time) error {
	if t.Status == StatusCancelled {
		return ErrCannotComplete
	}
	if t.Status == StatusDone {
		return ErrAlreadyDone
	}

	completed := now.UTC()
	t.Status = StatusDone
	t.CompletedAt = &completed
	t.UpdatedAt = completed
	return nil
}

func (t Task) Validate() error {
	if t.Title == "" {
		return ErrEmptyTitle
	}
	if !t.Status.Valid() {
		return ErrInvalidStatus
	}
	if !t.Priority.Valid() {
		return ErrInvalidPriority
	}
	if t.Status == StatusDone && t.CompletedAt == nil {
		return ErrCompletedAtNeeded
	}
	return nil
}
