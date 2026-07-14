package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/knowledge/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

const defaultListLimit = 10

type NoteStore interface {
	Save(ctx context.Context, note domain.Note) error
	ListRecent(ctx context.Context, userID ids.UserID, limit int32) ([]domain.Note, error)
	ListByTag(ctx context.Context, userID ids.UserID, tag string, limit int32) ([]domain.Note, error)
	Search(ctx context.Context, userID ids.UserID, query string, limit int32) ([]domain.Note, error)
	Delete(ctx context.Context, userID ids.UserID, noteID ids.NoteID) (domain.Note, error)
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type NoteDTO struct {
	ID        ids.NoteID
	Body      string
	Tags      []string
	CreatedAt time.Time
}

func ToNoteDTO(n domain.Note) NoteDTO {
	tags := n.Tags
	if tags == nil {
		tags = []string{}
	}
	return NoteDTO{ID: n.ID, Body: n.Body, Tags: tags, CreatedAt: n.CreatedAt}
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
	UserID ids.UserID
	Body   string
	Tags   []string
	Source events.Source
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
	note, err := domain.NewNote(in.UserID, body, tags, now)
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
