package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyTitle         = errors.New("task title is required")
	ErrCannotComplete     = errors.New("cancelled task cannot be completed")
	ErrAlreadyDone        = errors.New("task is already done")
	ErrInvalidStatus      = errors.New("invalid task status")
	ErrInvalidPriority    = errors.New("invalid task priority")
	ErrCompletedAtNeeded  = errors.New("completed task requires completed_at")
	ErrCannotCancelDone   = errors.New("done task cannot be cancelled")
	ErrAlreadyCancelled   = errors.New("task is already cancelled")
	ErrCannotEditTerminal = errors.New("done or cancelled task cannot be edited")
	ErrCannotReschedule   = errors.New("done or cancelled task cannot be rescheduled")
	ErrInvalidDuration    = errors.New("duration_minutes must be positive")
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
	ID              ids.TaskID
	UserID          ids.UserID
	Title           string
	Description     *string
	Status          Status
	Priority        Priority
	DueDate         *time.Time // дата реализации
	DurationMinutes *int       // оценка длительности в минутах
	Tags            []string   // хештеги без '#'
	ProjectIDs      []ids.ProjectID
	CompletedAt     *time.Time
	DeletedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
		Tags:       []string{},
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

func (t *Task) Cancel(now time.Time) error {
	if t.Status == StatusDone {
		return ErrCannotCancelDone
	}
	if t.Status == StatusCancelled {
		return ErrAlreadyCancelled
	}
	t.Status = StatusCancelled
	t.CompletedAt = nil
	t.UpdatedAt = now.UTC()
	return nil
}

func (t *Task) Reschedule(dueDate time.Time, now time.Time) error {
	if t.Status == StatusDone || t.Status == StatusCancelled {
		return ErrCannotReschedule
	}
	d := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, time.UTC)
	t.DueDate = &d
	t.UpdatedAt = now.UTC()
	return nil
}

type EditFields struct {
	Title            *string
	Description      *string
	ClearDescription bool
	Priority         *Priority
	DueDate          *time.Time
	ClearDueDate     bool
	DurationMinutes  *int
	ClearDuration    bool
	Tags             *[]string
}

func (t *Task) Edit(fields EditFields, now time.Time) error {
	if t.Status == StatusDone || t.Status == StatusCancelled {
		return ErrCannotEditTerminal
	}
	if fields.Title != nil {
		title := *fields.Title
		if title == "" {
			return ErrEmptyTitle
		}
		t.Title = title
	}
	if fields.ClearDescription {
		t.Description = nil
	} else if fields.Description != nil {
		t.Description = fields.Description
	}
	if fields.Priority != nil {
		if !fields.Priority.Valid() {
			return ErrInvalidPriority
		}
		t.Priority = *fields.Priority
	}
	if fields.ClearDueDate {
		t.DueDate = nil
	} else if fields.DueDate != nil {
		d := time.Date(fields.DueDate.Year(), fields.DueDate.Month(), fields.DueDate.Day(), 0, 0, 0, 0, time.UTC)
		t.DueDate = &d
	}
	if fields.ClearDuration {
		t.DurationMinutes = nil
	} else if fields.DurationMinutes != nil {
		if *fields.DurationMinutes <= 0 {
			return ErrInvalidDuration
		}
		mins := *fields.DurationMinutes
		t.DurationMinutes = &mins
	}
	if fields.Tags != nil {
		t.Tags = NormalizeTags(*fields.Tags)
	}
	t.UpdatedAt = now.UTC()
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
	if t.DurationMinutes != nil && *t.DurationMinutes <= 0 {
		return ErrInvalidDuration
	}
	return nil
}
