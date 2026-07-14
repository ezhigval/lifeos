package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/calendar/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type EventStore interface {
	Save(ctx context.Context, event domain.Event) error
	ListBetween(ctx context.Context, userID ids.UserID, from, to time.Time) ([]domain.Event, error)
}

type UserReader interface {
	Timezone(ctx context.Context, userID ids.UserID) (string, error)
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type EventDTO struct {
	ID       ids.EventID
	Title    string
	StartsAt time.Time
}

type CreateEvent struct {
	store      EventStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreateEvent(store EventStore, events EventLog, transactor Transactor) *CreateEvent {
	return &CreateEvent{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateEventInput struct {
	UserID   ids.UserID
	Title    string
	StartsAt time.Time
	Source   events.Source
}

func (uc *CreateEvent) Execute(ctx context.Context, in CreateEventInput) (EventDTO, error) {
	if in.UserID.IsZero() {
		return EventDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	event, err := domain.NewEvent(in.UserID, in.Title, in.StartsAt, now)
	if err != nil {
		return EventDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.Save(txCtx, event); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        event.UserID,
			AggregateType: "calendar_event",
			AggregateID:   event.ID.UUID(),
			EventType:     "CalendarEventCreated",
			Payload:       map[string]any{"title": event.Title, "starts_at": event.StartsAt},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return EventDTO{}, fmt.Errorf("create event: %w", err)
	}
	return EventDTO{ID: event.ID, Title: event.Title, StartsAt: event.StartsAt}, nil
}

type ListEventsToday struct {
	store EventStore
	users UserReader
	now   func() time.Time
}

func NewListEventsToday(store EventStore, users UserReader) *ListEventsToday {
	return &ListEventsToday{
		store: store,
		users: users,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

func (uc *ListEventsToday) Execute(ctx context.Context, userID ids.UserID) ([]EventDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load timezone: %w", err)
	}
	from, to, err := dayBounds(uc.now(), tz)
	if err != nil {
		return nil, err
	}
	items, err := uc.store.ListBetween(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	out := make([]EventDTO, 0, len(items))
	for _, item := range items {
		out = append(out, EventDTO{ID: item.ID, Title: item.Title, StartsAt: item.StartsAt})
	}
	return out, nil
}

func dayBounds(now time.Time, timezone string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := now.In(loc)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc).UTC()
	return start, start.Add(24 * time.Hour), nil
}
