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

const defaultStepsListLimit = 10

type StepStore interface {
	SaveSteps(ctx context.Context, log domain.StepLog) error
	GetLatestSteps(ctx context.Context, userID ids.UserID) (domain.StepLog, error)
	ListRecentSteps(ctx context.Context, userID ids.UserID, limit int32) ([]domain.StepLog, error)
}

type StepLogDTO struct {
	ID       ids.StepLogID
	Steps    int32
	LoggedAt time.Time
}

func ToStepLogDTO(l domain.StepLog) StepLogDTO {
	return StepLogDTO{ID: l.ID, Steps: l.Steps, LoggedAt: l.LoggedAt}
}

type RecordSteps struct {
	store      StepStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewRecordSteps(store StepStore, events EventLog, transactor Transactor) *RecordSteps {
	return &RecordSteps{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type RecordStepsInput struct {
	UserID   ids.UserID
	Steps    int32
	LoggedAt *time.Time
	Source   events.Source
}

func (uc *RecordSteps) Execute(ctx context.Context, in RecordStepsInput) (StepLogDTO, error) {
	if in.UserID.IsZero() {
		return StepLogDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	loggedAt := now
	if in.LoggedAt != nil {
		loggedAt = in.LoggedAt.UTC()
	}
	log, err := domain.NewStepLog(in.UserID, in.Steps, loggedAt, now)
	if err != nil {
		return StepLogDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.SaveSteps(txCtx, log); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        log.UserID,
			AggregateType: "step_log",
			AggregateID:   log.ID.UUID(),
			EventType:     "StepsRecorded",
			Payload:       map[string]any{"steps": log.Steps},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return StepLogDTO{}, fmt.Errorf("record steps: %w", err)
	}
	return ToStepLogDTO(log), nil
}

type GetLatestSteps struct {
	store StepStore
}

func NewGetLatestSteps(store StepStore) *GetLatestSteps {
	return &GetLatestSteps{store: store}
}

func (uc *GetLatestSteps) Execute(ctx context.Context, userID ids.UserID) (StepLogDTO, error) {
	if userID.IsZero() {
		return StepLogDTO{}, fmt.Errorf("user id is required")
	}
	log, err := uc.store.GetLatestSteps(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return StepLogDTO{}, err
	}
	if err != nil {
		return StepLogDTO{}, fmt.Errorf("get latest steps: %w", err)
	}
	return ToStepLogDTO(log), nil
}

type ListSteps struct {
	store StepStore
}

func NewListSteps(store StepStore) *ListSteps {
	return &ListSteps{store: store}
}

func (uc *ListSteps) Execute(ctx context.Context, userID ids.UserID) ([]StepLogDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	items, err := uc.store.ListRecentSteps(ctx, userID, defaultStepsListLimit)
	if err != nil {
		return nil, fmt.Errorf("list steps: %w", err)
	}
	out := make([]StepLogDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToStepLogDTO(item))
	}
	return out, nil
}
