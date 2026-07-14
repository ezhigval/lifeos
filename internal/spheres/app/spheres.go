package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/spheres/domain"
)

type SphereStore interface {
	Save(ctx context.Context, sphere domain.Sphere) error
	List(ctx context.Context, userID ids.UserID) ([]domain.Sphere, error)
	Get(ctx context.Context, userID ids.UserID, sphereID ids.SphereID) (domain.Sphere, error)
	FindByName(ctx context.Context, userID ids.UserID, name string) (domain.Sphere, error)
	Count(ctx context.Context, userID ids.UserID) (int32, error)
	Update(ctx context.Context, sphere domain.Sphere) error
	Delete(ctx context.Context, userID ids.UserID, sphereID ids.SphereID) (domain.Sphere, error)
	HasLinkedProjects(ctx context.Context, sphereID ids.SphereID) (bool, error)
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type SphereDTO struct {
	ID        ids.SphereID
	Name      string
	SortOrder int32
	CreatedAt time.Time
}

func ToSphereDTO(s domain.Sphere) SphereDTO {
	return SphereDTO{
		ID: s.ID, Name: s.Name, SortOrder: s.SortOrder, CreatedAt: s.CreatedAt,
	}
}

type CreateSphere struct {
	store      SphereStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreateSphere(store SphereStore, events EventLog, transactor Transactor) *CreateSphere {
	return &CreateSphere{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateSphereInput struct {
	UserID    ids.UserID
	Name      string
	SortOrder *int32
	Source    events.Source
}

func (uc *CreateSphere) Execute(ctx context.Context, in CreateSphereInput) (SphereDTO, error) {
	if in.UserID.IsZero() {
		return SphereDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	sortOrder := int32(0)
	if in.SortOrder != nil {
		sortOrder = *in.SortOrder
	} else {
		count, err := uc.store.Count(ctx, in.UserID)
		if err != nil {
			return SphereDTO{}, err
		}
		sortOrder = count
	}
	sphere, err := domain.NewSphere(in.UserID, in.Name, sortOrder, now)
	if err != nil {
		return SphereDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.Save(txCtx, sphere); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        sphere.UserID,
			AggregateType: "life_sphere",
			AggregateID:   sphere.ID.UUID(),
			EventType:     "SphereCreated",
			Payload:       map[string]any{"name": sphere.Name},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return SphereDTO{}, fmt.Errorf("create sphere: %w", err)
	}
	return ToSphereDTO(sphere), nil
}

type ListSpheres struct {
	store SphereStore
}

func NewListSpheres(store SphereStore) *ListSpheres {
	return &ListSpheres{store: store}
}

func (uc *ListSpheres) Execute(ctx context.Context, userID ids.UserID) ([]SphereDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	items, err := uc.store.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list spheres: %w", err)
	}
	out := make([]SphereDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToSphereDTO(item))
	}
	return out, nil
}

type FindSphereByName struct {
	store SphereStore
}

func NewFindSphereByName(store SphereStore) *FindSphereByName {
	return &FindSphereByName{store: store}
}

func (uc *FindSphereByName) Execute(ctx context.Context, userID ids.UserID, name string) (SphereDTO, error) {
	name = strings.TrimSpace(name)
	if userID.IsZero() || name == "" {
		return SphereDTO{}, fmt.Errorf("user id and name are required")
	}
	sphere, err := uc.store.FindByName(ctx, userID, name)
	if errors.Is(err, domain.ErrNotFound) {
		return SphereDTO{}, err
	}
	if err != nil {
		return SphereDTO{}, fmt.Errorf("find sphere: %w", err)
	}
	return ToSphereDTO(sphere), nil
}

type UpdateSphere struct {
	store      SphereStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewUpdateSphere(store SphereStore, events EventLog, transactor Transactor) *UpdateSphere {
	return &UpdateSphere{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type UpdateSphereInput struct {
	UserID    ids.UserID
	SphereID  ids.SphereID
	Name      string
	SortOrder int32
	Source    events.Source
}

func (uc *UpdateSphere) Execute(ctx context.Context, in UpdateSphereInput) (SphereDTO, error) {
	if in.UserID.IsZero() || in.SphereID.IsZero() {
		return SphereDTO{}, fmt.Errorf("user id and sphere id are required")
	}
	now := uc.now()
	sphere, err := uc.store.Get(ctx, in.UserID, in.SphereID)
	if errors.Is(err, domain.ErrNotFound) {
		return SphereDTO{}, err
	}
	if err != nil {
		return SphereDTO{}, fmt.Errorf("get sphere: %w", err)
	}
	if err := sphere.Rename(in.Name, in.SortOrder, now); err != nil {
		return SphereDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.Update(txCtx, sphere); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        sphere.UserID,
			AggregateType: "life_sphere",
			AggregateID:   sphere.ID.UUID(),
			EventType:     "SphereUpdated",
			Payload:       map[string]any{"name": sphere.Name},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return SphereDTO{}, fmt.Errorf("update sphere: %w", err)
	}
	return ToSphereDTO(sphere), nil
}

var ErrSphereNotFound = errors.New("sphere not found")

type DeleteSphere struct {
	store      SphereStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewDeleteSphere(store SphereStore, events EventLog, transactor Transactor) *DeleteSphere {
	return &DeleteSphere{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type DeleteSphereInput struct {
	UserID   ids.UserID
	SphereID ids.SphereID
	Source   events.Source
}

func (uc *DeleteSphere) Execute(ctx context.Context, in DeleteSphereInput) (SphereDTO, error) {
	if in.UserID.IsZero() || in.SphereID.IsZero() {
		return SphereDTO{}, fmt.Errorf("user id and sphere id are required")
	}
	linked, err := uc.store.HasLinkedProjects(ctx, in.SphereID)
	if err != nil {
		return SphereDTO{}, err
	}
	if linked {
		return SphereDTO{}, domain.ErrHasProjects
	}
	now := uc.now()
	var deleted SphereDTO
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		sphere, err := uc.store.Delete(txCtx, in.UserID, in.SphereID)
		if errors.Is(err, domain.ErrNotFound) {
			return ErrSphereNotFound
		}
		if err != nil {
			return err
		}
		deleted = ToSphereDTO(sphere)
		return uc.events.Append(txCtx, events.Record{
			UserID:        sphere.UserID,
			AggregateType: "life_sphere",
			AggregateID:   sphere.ID.UUID(),
			EventType:     "SphereDeleted",
			Payload:       map[string]any{"name": sphere.Name},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		if errors.Is(err, ErrSphereNotFound) {
			return SphereDTO{}, ErrSphereNotFound
		}
		return SphereDTO{}, fmt.Errorf("delete sphere: %w", err)
	}
	return deleted, nil
}

type EnsureDefaultSpheres struct {
	store SphereStore
	now   func() time.Time
}

func NewEnsureDefaultSpheres(store SphereStore) *EnsureDefaultSpheres {
	return &EnsureDefaultSpheres{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

func (uc *EnsureDefaultSpheres) Execute(ctx context.Context, userID ids.UserID) error {
	if userID.IsZero() {
		return fmt.Errorf("user id is required")
	}
	count, err := uc.store.Count(ctx, userID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := uc.now()
	for i, name := range domain.DefaultSphereNames {
		sphere, err := domain.NewSphere(userID, name, int32(i), now)
		if err != nil {
			return err
		}
		if err := uc.store.Save(ctx, sphere); err != nil {
			return fmt.Errorf("seed sphere %q: %w", name, err)
		}
	}
	return nil
}
