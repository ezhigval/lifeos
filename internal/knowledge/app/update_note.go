package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/knowledge/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type NoteReader interface {
	GetByID(ctx context.Context, userID ids.UserID, noteID ids.NoteID) (domain.Note, error)
}

type NoteWriter interface {
	UpdateBody(ctx context.Context, userID ids.UserID, noteID ids.NoteID, body string, now time.Time) (domain.Note, error)
}

type GetNote struct {
	notes NoteReader
}

func NewGetNote(notes NoteReader) *GetNote {
	return &GetNote{notes: notes}
}

func (uc *GetNote) Execute(ctx context.Context, userID ids.UserID, noteID ids.NoteID) (NoteDTO, error) {
	if userID.IsZero() || noteID.IsZero() {
		return NoteDTO{}, fmt.Errorf("user id and note id are required")
	}
	note, err := uc.notes.GetByID(ctx, userID, noteID)
	if errors.Is(err, domain.ErrNotFound) {
		return NoteDTO{}, ErrNoteNotFound
	}
	if err != nil {
		return NoteDTO{}, err
	}
	return ToNoteDTO(note), nil
}

type UpdateNote struct {
	notes NoteWriter
	now   func() time.Time
}

func NewUpdateNote(notes NoteWriter) *UpdateNote {
	return &UpdateNote{notes: notes, now: func() time.Time { return time.Now().UTC() }}
}

func (uc *UpdateNote) Execute(ctx context.Context, userID ids.UserID, noteID ids.NoteID, body string) (NoteDTO, error) {
	if userID.IsZero() || noteID.IsZero() {
		return NoteDTO{}, fmt.Errorf("user id and note id are required")
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return NoteDTO{}, domain.ErrEmptyBody
	}
	note, err := uc.notes.UpdateBody(ctx, userID, noteID, body, uc.now())
	if errors.Is(err, domain.ErrNotFound) {
		return NoteDTO{}, ErrNoteNotFound
	}
	if err != nil {
		return NoteDTO{}, err
	}
	return ToNoteDTO(note), nil
}
