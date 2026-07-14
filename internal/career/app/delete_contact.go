package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/career/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var ErrContactNotFound = errors.New("contact not found")

type DeleteContact struct {
	store      ContactStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewDeleteContact(store ContactStore, events EventLog, transactor Transactor) *DeleteContact {
	return &DeleteContact{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type DeleteContactInput struct {
	UserID    ids.UserID
	ContactID ids.ContactID
	Source    events.Source
}

func (uc *DeleteContact) Execute(ctx context.Context, in DeleteContactInput) (ContactDTO, error) {
	if in.UserID.IsZero() || in.ContactID.IsZero() {
		return ContactDTO{}, fmt.Errorf("user id and contact id are required")
	}
	now := uc.now()
	var deleted ContactDTO
	err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		contact, err := uc.store.Delete(txCtx, in.UserID, in.ContactID)
		if errors.Is(err, domain.ErrNotFound) {
			return ErrContactNotFound
		}
		if err != nil {
			return err
		}
		deleted = ToContactDTO(contact)
		return uc.events.Append(txCtx, events.Record{
			UserID:        contact.UserID,
			AggregateType: "contact",
			AggregateID:   contact.ID.UUID(),
			EventType:     "ContactDeleted",
			Payload:       map[string]any{"name": contact.Name},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		if errors.Is(err, ErrContactNotFound) {
			return ContactDTO{}, ErrContactNotFound
		}
		return ContactDTO{}, fmt.Errorf("delete contact: %w", err)
	}
	return deleted, nil
}
