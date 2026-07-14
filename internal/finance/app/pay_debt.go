package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type DebtPaymentStore interface {
	GetByID(ctx context.Context, userID ids.UserID, debtID ids.DebtID) (domain.Debt, error)
	FindOpenByCreditor(ctx context.Context, userID ids.UserID, creditor string) (domain.Debt, error)
	UpdateDebt(ctx context.Context, debt domain.Debt) error
}

type PayDebt struct {
	debts      DebtPaymentStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewPayDebt(debts DebtPaymentStore, events EventLog, transactor Transactor) *PayDebt {
	return &PayDebt{
		debts:      debts,
		events:     events,
		transactor: transactor,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

type PayDebtInput struct {
	UserID      ids.UserID
	DebtID      ids.DebtID
	Creditor    string
	AmountCents int64
	// Regular marks an installment payment and advances next_payment_date.
	Regular bool
	Source  events.Source
}

func (uc *PayDebt) Execute(ctx context.Context, in PayDebtInput) (DebtDTO, error) {
	if in.UserID.IsZero() {
		return DebtDTO{}, fmt.Errorf("user id is required")
	}
	if in.AmountCents <= 0 {
		return DebtDTO{}, domain.ErrInvalidAmount
	}

	var debt domain.Debt
	var err error
	switch {
	case !in.DebtID.IsZero():
		debt, err = uc.debts.GetByID(ctx, in.UserID, in.DebtID)
	case in.Creditor != "":
		debt, err = uc.debts.FindOpenByCreditor(ctx, in.UserID, in.Creditor)
	default:
		return DebtDTO{}, fmt.Errorf("debt id or creditor is required")
	}
	if errors.Is(err, domain.ErrDebtNotFound) {
		return DebtDTO{}, err
	}
	if err != nil {
		return DebtDTO{}, fmt.Errorf("find debt: %w", err)
	}

	now := uc.now()
	if err := debt.Pay(in.AmountCents); err != nil {
		return DebtDTO{}, err
	}
	if in.Regular {
		debt.AdvanceInstallment(now)
	}

	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.debts.UpdateDebt(txCtx, debt); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        debt.UserID,
			AggregateType: "debt",
			AggregateID:   debt.ID.UUID(),
			EventType:     "DebtPaymentRecorded",
			Payload: map[string]any{
				"creditor":   debt.Creditor,
				"paid_cents": in.AmountCents,
				"remaining":  debt.RemainingCents(),
				"status":     debt.Status,
				"regular":    in.Regular,
			},
			Source:     in.Source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return DebtDTO{}, fmt.Errorf("pay debt: %w", err)
	}
	return ToDebtDTO(debt), nil
}
