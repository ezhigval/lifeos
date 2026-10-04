package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrEmptyBody = errors.New("note body is required")
var ErrNotFound = errors.New("note not found")
var ErrInvalidTarget = errors.New("invalid note target")

// TargetType — тип сущности, к которой привязана заметка (TASK-011 item 5:
// заметка из задачи/события/напоминания видна в общем списке с бейджем цели).
type TargetType string

const (
	TargetTask     TargetType = "task"
	TargetEvent    TargetType = "event"
	TargetReminder TargetType = "reminder"
)

func (t TargetType) Valid() bool {
	switch t {
	case TargetTask, TargetEvent, TargetReminder:
		return true
	}
	return false
}

type Note struct {
	ID         ids.NoteID
	UserID     ids.UserID
	Body       string
	Tags       []string
	TargetType *TargetType
	TargetID   *uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewNote(userID ids.UserID, body string, tags []string, now time.Time) (Note, error) {
	return NewNoteWithTarget(userID, body, tags, nil, nil, now)
}

func NewNoteWithTarget(userID ids.UserID, body string, tags []string, targetType *TargetType, targetID *uuid.UUID, now time.Time) (Note, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return Note{}, ErrEmptyBody
	}
	if targetType != nil && !targetType.Valid() {
		return Note{}, fmt.Errorf("%w: %q", ErrInvalidTarget, string(*targetType))
	}
	if targetType != nil && targetID == nil {
		return Note{}, fmt.Errorf("%w: target id is required", ErrInvalidTarget)
	}
	now = now.UTC()
	return Note{
		ID:         ids.NewNoteID(),
		UserID:     userID,
		Body:       body,
		Tags:       NormalizeTags(tags),
		TargetType: targetType,
		TargetID:   targetID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
