package domain

import (
	"errors"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrEmptyTitle = errors.New("event title is required")

type Event struct {
	ID        ids.EventID
	UserID    ids.UserID
	Title     string
	StartsAt  time.Time
	EndsAt    *time.Time
	CreatedAt time.Time
}

func NewEvent(userID ids.UserID, title string, startsAt time.Time, now time.Time) (Event, error) {
	if title == "" {
		return Event{}, ErrEmptyTitle
	}
	return Event{
		ID:        ids.NewEventID(),
		UserID:    userID,
		Title:     title,
		StartsAt:  startsAt.UTC(),
		CreatedAt: now.UTC(),
	}, nil
}
