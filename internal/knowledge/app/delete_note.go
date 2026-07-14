package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/knowledge/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrNoteNotFound = errors.New("note not found")

type DeleteNote struct {
	notes      NoteStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewDeleteNote(notes NoteStore, events EventLog, transactor Transactor) *DeleteNote {
	return &DeleteNote{
		notes: notes, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type DeleteNoteInput struct {
	UserID ids.UserID
	NoteID ids.NoteID
	Source events.Source
}

func (uc *DeleteNote) Execute(ctx context.Context, in DeleteNoteInput) (NoteDTO, error) {
	if in.UserID.IsZero() || in.NoteID.IsZero() {
		return NoteDTO{}, fmt.Errorf("user id and note id are required")
	}
	now := uc.now()
	var deleted NoteDTO
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		note, err := uc.notes.Delete(txCtx, in.UserID, in.NoteID)
		if errors.Is(err, domain.ErrNotFound) {
			return ErrNoteNotFound
		}
		if err != nil {
			return err
		}
		deleted = ToNoteDTO(note)
		return uc.events.Append(txCtx, events.Record{
			UserID:        note.UserID,
			AggregateType: "note",
			AggregateID:   note.ID.UUID(),
			EventType:     "NoteDeleted",
			Payload:       map[string]any{"body": note.Body},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			return NoteDTO{}, ErrNoteNotFound
		}
		return NoteDTO{}, fmt.Errorf("delete note: %w", err)
	}
	return deleted, nil
}
