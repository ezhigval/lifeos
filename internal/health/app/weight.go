package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/health/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

const defaultWeightListLimit = 10

type WeightStore interface {
	Save(ctx context.Context, log domain.WeightLog) error
	GetLatest(ctx context.Context, userID ids.UserID) (domain.WeightLog, error)
	ListRecent(ctx context.Context, userID ids.UserID, limit int32) ([]domain.WeightLog, error)
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type WeightLogDTO struct {
	ID       ids.WeightLogID
	WeightKg float64
	LoggedAt time.Time
}

func ToWeightLogDTO(l domain.WeightLog) WeightLogDTO {
	return WeightLogDTO{ID: l.ID, WeightKg: l.WeightKg, LoggedAt: l.LoggedAt}
}

type RecordWeight struct {
	store      WeightStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewRecordWeight(store WeightStore, events EventLog, transactor Transactor) *RecordWeight {
	return &RecordWeight{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type RecordWeightInput struct {
	UserID   ids.UserID
	WeightKg float64
	LoggedAt *time.Time
	Source   events.Source
}

func (uc *RecordWeight) Execute(ctx context.Context, in RecordWeightInput) (WeightLogDTO, error) {
	if in.UserID.IsZero() {
		return WeightLogDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	loggedAt := now
	if in.LoggedAt != nil {
		loggedAt = in.LoggedAt.UTC()
	}
	log, err := domain.NewWeightLog(in.UserID, in.WeightKg, loggedAt, now)
	if err != nil {
		return WeightLogDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.Save(txCtx, log); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        log.UserID,
			AggregateType: "weight_log",
			AggregateID:   log.ID.UUID(),
			EventType:     "WeightRecorded",
			Payload:       map[string]any{"weight_kg": log.WeightKg},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return WeightLogDTO{}, fmt.Errorf("record weight: %w", err)
	}
	return ToWeightLogDTO(log), nil
}

type GetLatestWeight struct {
	store WeightStore
}

func NewGetLatestWeight(store WeightStore) *GetLatestWeight {
	return &GetLatestWeight{store: store}
}

func (uc *GetLatestWeight) Execute(ctx context.Context, userID ids.UserID) (WeightLogDTO, error) {
	if userID.IsZero() {
		return WeightLogDTO{}, fmt.Errorf("user id is required")
	}
	log, err := uc.store.GetLatest(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return WeightLogDTO{}, err
	}
	if err != nil {
		return WeightLogDTO{}, fmt.Errorf("get latest weight: %w", err)
	}
	return ToWeightLogDTO(log), nil
}

type ListWeights struct {
	store WeightStore
}

func NewListWeights(store WeightStore) *ListWeights {
	return &ListWeights{store: store}
}

func (uc *ListWeights) Execute(ctx context.Context, userID ids.UserID) ([]WeightLogDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	items, err := uc.store.ListRecent(ctx, userID, defaultWeightListLimit)
	if err != nil {
		return nil, fmt.Errorf("list weights: %w", err)
	}
	out := make([]WeightLogDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToWeightLogDTO(item))
	}
	return out, nil
}
