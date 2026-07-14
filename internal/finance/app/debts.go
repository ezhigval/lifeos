package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type DebtStore interface {
	SaveDebt(ctx context.Context, debt domain.Debt) error
	ListOpen(ctx context.Context, userID ids.UserID) ([]domain.Debt, error)
}

type DebtDTO struct {
	ID                  ids.DebtID
	Creditor            string
	AmountCents         int64
	PaidCents           int64
	RemainingCents      int64
	Currency            string
	DueDate             *time.Time
	InstallmentCents    int64
	InstallmentInterval string
	NextPaymentDate     *time.Time
}

func ToDebtDTO(d domain.Debt) DebtDTO {
	interval := d.InstallmentInterval
	if interval == "" {
		interval = "none"
	}
	return DebtDTO{
		ID:                  d.ID,
		Creditor:            d.Creditor,
		AmountCents:         d.AmountCents,
		PaidCents:           d.PaidCents,
		RemainingCents:      d.RemainingCents(),
		Currency:            "RUB",
		DueDate:             d.DueDate,
		InstallmentCents:    d.InstallmentCents,
		InstallmentInterval: interval,
		NextPaymentDate:     d.NextPaymentDate,
	}
}

type CreateDebt struct {
	debts      DebtStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreateDebt(debts DebtStore, events EventLog, transactor Transactor) *CreateDebt {
	return &CreateDebt{
		debts:      debts,
		events:     events,
		transactor: transactor,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

type CreateDebtInput struct {
	UserID              ids.UserID
	Creditor            string
	AmountCents         int64
	DueDate             *time.Time
	InstallmentCents    int64
	InstallmentInterval string
	NextPaymentDate     *time.Time
	Source              events.Source
}

func (uc *CreateDebt) Execute(ctx context.Context, in CreateDebtInput) (DebtDTO, error) {
	if in.UserID.IsZero() {
		return DebtDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	debt, err := domain.NewDebt(in.UserID, in.Creditor, in.AmountCents, in.DueDate, now)
	if err != nil {
		return DebtDTO{}, err
	}
	if in.InstallmentCents > 0 {
		debt.InstallmentCents = in.InstallmentCents
		interval := in.InstallmentInterval
		if interval == "" {
			interval = "monthly"
		}
		debt.InstallmentInterval = interval
		if in.NextPaymentDate != nil {
			debt.NextPaymentDate = in.NextPaymentDate
		} else if in.DueDate != nil {
			debt.NextPaymentDate = in.DueDate
		}
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.debts.SaveDebt(txCtx, debt); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        debt.UserID,
			AggregateType: "debt",
			AggregateID:   debt.ID.UUID(),
			EventType:     "DebtCreated",
			Payload: map[string]any{
				"creditor":     debt.Creditor,
				"amount_cents": debt.AmountCents,
				"due_date":     debt.DueDate,
			},
			Source:     in.Source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return DebtDTO{}, fmt.Errorf("create debt: %w", err)
	}
	return ToDebtDTO(debt), nil
}

type ListDebts struct {
	debts DebtStore
}

func NewListDebts(debts DebtStore) *ListDebts {
	return &ListDebts{debts: debts}
}

func (uc *ListDebts) Execute(ctx context.Context, userID ids.UserID) ([]DebtDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	items, err := uc.debts.ListOpen(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list debts: %w", err)
	}
	out := make([]DebtDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToDebtDTO(item))
	}
	return out, nil
}
