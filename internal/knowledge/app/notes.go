package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/valentinezhov/lifeos/internal/knowledge/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

const defaultListLimit = 10

type NoteStore interface {
	Save(ctx context.Context, note domain.Note) error
	GetByID(ctx context.Context, userID ids.UserID, noteID ids.NoteID) (domain.Note, error)
	UpdateBody(ctx context.Context, userID ids.UserID, noteID ids.NoteID, body string, now time.Time) (domain.Note, error)
	ListRecent(ctx context.Context, userID ids.UserID, limit int32) ([]domain.Note, error)
	ListByTag(ctx context.Context, userID ids.UserID, tag string, limit int32) ([]domain.Note, error)
	Search(ctx context.Context, userID ids.UserID, query string, limit int32) ([]domain.Note, error)
	ListCreatedBetween(ctx context.Context, userID ids.UserID, from, to time.Time) ([]domain.Note, error)
	ListByTarget(ctx context.Context, userID ids.UserID, targetType domain.TargetType, targetID uuid.UUID) ([]domain.Note, error)
	Delete(ctx context.Context, userID ids.UserID, noteID ids.NoteID) (domain.Note, error)
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type NoteDTO struct {
	ID         ids.NoteID
	Body       string
	Tags       []string
	TargetType string
	TargetID   string
	CreatedAt  time.Time
}

func ToNoteDTO(n domain.Note) NoteDTO {
	tags := n.Tags
	if tags == nil {
		tags = []string{}
	}
	dto := NoteDTO{ID: n.ID, Body: n.Body, Tags: tags, CreatedAt: n.CreatedAt}
	if n.TargetType != nil {
		dto.TargetType = string(*n.TargetType)
	}
	if n.TargetID != nil {
		dto.TargetID = n.TargetID.String()
	}
	return dto
}

type CreateNote struct {
	notes      NoteStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreateNote(notes NoteStore, events EventLog, transactor Transactor) *CreateNote {
	return &CreateNote{
		notes: notes, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateNoteInput struct {
	UserID     ids.UserID
	Body       string
	Tags       []string
	Source     events.Source
	TargetType string // "" | task | event | reminder (TASK-011 item 5)
	TargetID   string // uuid of the linked entity
}

func (uc *CreateNote) Execute(ctx context.Context, in CreateNoteInput) (NoteDTO, error) {
	if in.UserID.IsZero() {
		return NoteDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	body := strings.TrimSpace(in.Body)
	tags := domain.NormalizeTags(in.Tags)
	if len(tags) == 0 {
		body, tags = domain.ExtractHashtags(body)
	}
	var targetType *domain.TargetType
	var targetID *uuid.UUID
	if tt := strings.TrimSpace(in.TargetType); tt != "" {
		parsed := domain.TargetType(tt)
		if !parsed.Valid() {
			return NoteDTO{}, fmt.Errorf("invalid target type %q", tt)
		}
		id, err := uuid.Parse(strings.TrimSpace(in.TargetID))
		if err != nil {
			return NoteDTO{}, fmt.Errorf("target_id must be a valid uuid when target_type is set")
		}
		targetType, targetID = &parsed, &id
	} else if strings.TrimSpace(in.TargetID) != "" {
		return NoteDTO{}, fmt.Errorf("target_type is required when target_id is set")
	}
	note, err := domain.NewNoteWithTarget(in.UserID, body, tags, targetType, targetID, now)
	if err != nil {
		return NoteDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.notes.Save(txCtx, note); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        note.UserID,
			AggregateType: "note",
			AggregateID:   note.ID.UUID(),
			EventType:     "NoteCreated",
			Payload:       map[string]any{"body": note.Body, "tags": note.Tags},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return NoteDTO{}, fmt.Errorf("create note: %w", err)
	}
	return ToNoteDTO(note), nil
}

type ListNotes struct {
	notes NoteStore
}

func NewListNotes(notes NoteStore) *ListNotes {
	return &ListNotes{notes: notes}
}

type ListNotesInput struct {
	UserID ids.UserID
	Tag    string
}

func (uc *ListNotes) Execute(ctx context.Context, in ListNotesInput) ([]NoteDTO, error) {
	if in.UserID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	var items []domain.Note
	var err error
	if tag := strings.TrimSpace(strings.TrimPrefix(in.Tag, "#")); tag != "" {
		items, err = uc.notes.ListByTag(ctx, in.UserID, strings.ToLower(tag), defaultListLimit)
	} else {
		items, err = uc.notes.ListRecent(ctx, in.UserID, defaultListLimit)
	}
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	out := make([]NoteDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToNoteDTO(item))
	}
	return out, nil
}

type SearchNotes struct {
	notes NoteStore
}

// ListNotesBetween returns notes created in [from, to) — used by the calendar
// agenda view so notes surface on their creation day.
type ListNotesBetween struct {
	notes NoteStore
}

func NewListNotesBetween(notes NoteStore) *ListNotesBetween {
	return &ListNotesBetween{notes: notes}
}

func (uc *ListNotesBetween) Execute(ctx context.Context, userID ids.UserID, from, to time.Time) ([]NoteDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("invalid date range")
	}
	items, err := uc.notes.ListCreatedBetween(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list notes between: %w", err)
	}
	out := make([]NoteDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToNoteDTO(item))
	}
	return out, nil
}

// ListNotesByTarget returns notes linked to a specific entity (task/event/
// reminder) — the reverse side of the TASK-011 item 5 two-way sync: opening a
// task shows its notes, and a note card shows a badge with its target.
type ListNotesByTarget struct {
	notes NoteStore
}

func NewListNotesByTarget(notes NoteStore) *ListNotesByTarget {
	return &ListNotesByTarget{notes: notes}
}

func (uc *ListNotesByTarget) Execute(ctx context.Context, userID ids.UserID, targetType, targetID string) ([]NoteDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	parsed := domain.TargetType(strings.TrimSpace(targetType))
	if !parsed.Valid() {
		return nil, fmt.Errorf("invalid target type %q", targetType)
	}
	id, err := uuid.Parse(strings.TrimSpace(targetID))
	if err != nil {
		return nil, fmt.Errorf("target_id must be a valid uuid")
	}
	items, err := uc.notes.ListByTarget(ctx, userID, parsed, id)
	if err != nil {
		return nil, fmt.Errorf("list notes by target: %w", err)
	}
	out := make([]NoteDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToNoteDTO(item))
	}
	return out, nil
}

func NewSearchNotes(notes NoteStore) *SearchNotes {
	return &SearchNotes{notes: notes}
}

type SearchNotesInput struct {
	UserID ids.UserID
	Query  string
}

func (uc *SearchNotes) Execute(ctx context.Context, in SearchNotesInput) ([]NoteDTO, error) {
	if in.UserID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}
	items, err := uc.notes.Search(ctx, in.UserID, query, defaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}
	out := make([]NoteDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToNoteDTO(item))
	}
	return out, nil
}
