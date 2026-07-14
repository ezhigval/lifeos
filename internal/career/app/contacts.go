package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/career/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

const defaultListLimit = 10

type ContactStore interface {
	Save(ctx context.Context, contact domain.Contact) error
	ListRecent(ctx context.Context, userID ids.UserID, limit int32) ([]domain.Contact, error)
	Search(ctx context.Context, userID ids.UserID, query string, limit int32) ([]domain.Contact, error)
	Delete(ctx context.Context, userID ids.UserID, contactID ids.ContactID) (domain.Contact, error)
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ContactDTO struct {
	ID        ids.ContactID
	Name      string
	Company   string
	Role      string
	Notes     string
	CreatedAt time.Time
}

func ToContactDTO(c domain.Contact) ContactDTO {
	return ContactDTO{
		ID: c.ID, Name: c.Name, Company: c.Company, Role: c.Role, Notes: c.Notes, CreatedAt: c.CreatedAt,
	}
}

type CreateContact struct {
	store      ContactStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreateContact(store ContactStore, events EventLog, transactor Transactor) *CreateContact {
	return &CreateContact{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateContactInput struct {
	UserID  ids.UserID
	Name    string
	Company string
	Role    string
	Notes   string
	Source  events.Source
}

func (uc *CreateContact) Execute(ctx context.Context, in CreateContactInput) (ContactDTO, error) {
	if in.UserID.IsZero() {
		return ContactDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	contact, err := domain.NewContact(in.UserID, in.Name, in.Company, in.Role, in.Notes, now)
	if err != nil {
		return ContactDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.Save(txCtx, contact); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        contact.UserID,
			AggregateType: "contact",
			AggregateID:   contact.ID.UUID(),
			EventType:     "ContactCreated",
			Payload: map[string]any{
				"name": contact.Name, "company": contact.Company, "role": contact.Role,
			},
			Source:     in.Source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return ContactDTO{}, fmt.Errorf("create contact: %w", err)
	}
	return ToContactDTO(contact), nil
}

type ListContacts struct {
	store ContactStore
}

func NewListContacts(store ContactStore) *ListContacts {
	return &ListContacts{store: store}
}

func (uc *ListContacts) Execute(ctx context.Context, userID ids.UserID) ([]ContactDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	items, err := uc.store.ListRecent(ctx, userID, defaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	out := make([]ContactDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToContactDTO(item))
	}
	return out, nil
}

type SearchContacts struct {
	store ContactStore
}

func NewSearchContacts(store ContactStore) *SearchContacts {
	return &SearchContacts{store: store}
}

type SearchContactsInput struct {
	UserID ids.UserID
	Query  string
}

func (uc *SearchContacts) Execute(ctx context.Context, in SearchContactsInput) ([]ContactDTO, error) {
	if in.UserID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}
	items, err := uc.store.Search(ctx, in.UserID, query, defaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("search contacts: %w", err)
	}
	out := make([]ContactDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToContactDTO(item))
	}
	return out, nil
}
