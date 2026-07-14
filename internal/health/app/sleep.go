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

const defaultSleepListLimit = 10

type SleepStore interface {
	SaveSleep(ctx context.Context, log domain.SleepLog) error
	GetLatestSleep(ctx context.Context, userID ids.UserID) (domain.SleepLog, error)
	ListRecentSleep(ctx context.Context, userID ids.UserID, limit int32) ([]domain.SleepLog, error)
}

type SleepLogDTO struct {
	ID              ids.SleepLogID
	DurationMinutes int32
	LoggedAt        time.Time
}

func (d SleepLogDTO) DurationHours() float64 {
	return domain.DurationHours(d.DurationMinutes)
}

func ToSleepLogDTO(l domain.SleepLog) SleepLogDTO {
	return SleepLogDTO{
		ID:              l.ID,
		DurationMinutes: l.DurationMinutes,
		LoggedAt:        l.LoggedAt,
	}
}

type RecordSleep struct {
	store      SleepStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewRecordSleep(store SleepStore, events EventLog, transactor Transactor) *RecordSleep {
	return &RecordSleep{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type RecordSleepInput struct {
	UserID          ids.UserID
	DurationMinutes int32
	LoggedAt        *time.Time
	Source          events.Source
}

func (uc *RecordSleep) Execute(ctx context.Context, in RecordSleepInput) (SleepLogDTO, error) {
	if in.UserID.IsZero() {
		return SleepLogDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	loggedAt := now
	if in.LoggedAt != nil {
		loggedAt = in.LoggedAt.UTC()
	}
	log, err := domain.NewSleepLog(in.UserID, in.DurationMinutes, loggedAt, now)
	if err != nil {
		return SleepLogDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.SaveSleep(txCtx, log); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        log.UserID,
			AggregateType: "sleep_log",
			AggregateID:   log.ID.UUID(),
			EventType:     "SleepRecorded",
			Payload:       map[string]any{"duration_minutes": log.DurationMinutes},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return SleepLogDTO{}, fmt.Errorf("record sleep: %w", err)
	}
	return ToSleepLogDTO(log), nil
}

type GetLatestSleep struct {
	store SleepStore
}

func NewGetLatestSleep(store SleepStore) *GetLatestSleep {
	return &GetLatestSleep{store: store}
}

func (uc *GetLatestSleep) Execute(ctx context.Context, userID ids.UserID) (SleepLogDTO, error) {
	if userID.IsZero() {
		return SleepLogDTO{}, fmt.Errorf("user id is required")
	}
	log, err := uc.store.GetLatestSleep(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return SleepLogDTO{}, err
	}
	if err != nil {
		return SleepLogDTO{}, fmt.Errorf("get latest sleep: %w", err)
	}
	return ToSleepLogDTO(log), nil
}

type ListSleep struct {
	store SleepStore
}

func NewListSleep(store SleepStore) *ListSleep {
	return &ListSleep{store: store}
}

func (uc *ListSleep) Execute(ctx context.Context, userID ids.UserID) ([]SleepLogDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	items, err := uc.store.ListRecentSleep(ctx, userID, defaultSleepListLimit)
	if err != nil {
		return nil, fmt.Errorf("list sleep: %w", err)
	}
	out := make([]SleepLogDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToSleepLogDTO(item))
	}
	return out, nil
}
