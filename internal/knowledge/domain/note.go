package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrEmptyBody = errors.New("note body is required")
var ErrNotFound = errors.New("note not found")

type Note struct {
	ID        ids.NoteID
	UserID    ids.UserID
	Body      string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewNote(userID ids.UserID, body string, tags []string, now time.Time) (Note, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return Note{}, ErrEmptyBody
	}
	now = now.UTC()
	return Note{
		ID:        ids.NewNoteID(),
		UserID:    userID,
		Body:      body,
		Tags:      NormalizeTags(tags),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
