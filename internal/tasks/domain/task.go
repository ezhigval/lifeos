package domain

import (
	"errors"
	"strings"
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
	ErrCannotArchiveDone  = errors.New("done task cannot be archived")
	ErrAlreadyDeleted     = errors.New("task is already deleted")
	ErrInvalidKind        = errors.New("invalid task kind")
	ErrCannotReopen       = errors.New("only done tasks can be reopened")
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

type Kind string

const (
	KindTask     Kind = "task"
	KindReminder Kind = "reminder"
	KindMeeting  Kind = "meeting"
)

type Task struct {
	ID              ids.TaskID
	UserID          ids.UserID
	Title           string
	Description     *string
	Status          Status
	Priority        Priority
	Kind            Kind
	Address         *string
	NoteID          *ids.NoteID
	DueDate         *time.Time // дата реализации / встречи / пуша
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
		Kind:       KindTask,
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

func (k Kind) Valid() bool {
	switch k {
	case KindTask, KindReminder, KindMeeting:
		return true
	default:
		return false
	}
}

func (t Task) KindOrDefault() Kind {
	if t.Kind == "" {
		return KindTask
	}
	return t.Kind
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

// Reopen returns a done task to the active queue.
func (t *Task) Reopen(now time.Time) error {
	if t.Status != StatusDone {
		return ErrCannotReopen
	}
	t.Status = StatusTodo
	t.CompletedAt = nil
	t.UpdatedAt = now.UTC()
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

// ApplyEdit updates title, priority, due date and optional description (full replace).
func (t *Task) ApplyEdit(title string, priority Priority, dueDate *time.Time, description *string, now time.Time) error {
	if t.Status == StatusDone || t.Status == StatusCancelled {
		return ErrCannotEditTerminal
	}
	if title == "" {
		return ErrEmptyTitle
	}
	if !priority.Valid() {
		return ErrInvalidPriority
	}
	t.Title = title
	t.Priority = priority
	t.DueDate = dueDate
	t.Description = description
	t.UpdatedAt = now.UTC()
	return t.Validate()
}

// Archive marks the task as cancelled (soft archive, still in DB).
func (t *Task) Archive(now time.Time) error {
	if t.DeletedAt != nil {
		return ErrAlreadyDeleted
	}
	if t.Status == StatusCancelled {
		return ErrAlreadyCancelled
	}
	if t.Status == StatusDone {
		return ErrCannotArchiveDone
	}
	t.Status = StatusCancelled
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
	Kind             *Kind
	Address          *string
	ClearAddress     bool
	NoteID           *ids.NoteID
	ClearNoteID      bool
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
	if fields.Kind != nil {
		if !fields.Kind.Valid() {
			return ErrInvalidKind
		}
		t.Kind = *fields.Kind
	}
	if fields.ClearAddress {
		t.Address = nil
	} else if fields.Address != nil {
		addr := strings.TrimSpace(*fields.Address)
		if addr == "" {
			t.Address = nil
		} else {
			t.Address = &addr
		}
	}
	if fields.ClearNoteID {
		t.NoteID = nil
	} else if fields.NoteID != nil {
		id := *fields.NoteID
		t.NoteID = &id
	}
	t.UpdatedAt = now.UTC()
	return nil
}

// SoftDelete sets deleted_at (hidden from lists).
func (t *Task) SoftDelete(now time.Time) error {
	if t.DeletedAt != nil {
		return ErrAlreadyDeleted
	}
	deleted := now.UTC()
	t.DeletedAt = &deleted
	t.UpdatedAt = deleted
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
	if t.Kind == "" {
		// legacy rows / in-memory defaults before Kind was set
	} else if !t.Kind.Valid() {
		return ErrInvalidKind
	}
	if t.Status == StatusDone && t.CompletedAt == nil {
		return ErrCompletedAtNeeded
	}
	if t.DurationMinutes != nil && *t.DurationMinutes <= 0 {
		return ErrInvalidDuration
	}
	return nil
}
