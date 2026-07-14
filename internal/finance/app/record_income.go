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

const defaultIncomeCategory = "Доход"

type CategoryStore interface {
	GetByName(ctx context.Context, userID ids.UserID, name string, kind domain.Kind) (domain.Category, error)
	SaveCategory(ctx context.Context, cat domain.Category) error
}

type TransactionStore interface {
	SaveTransaction(ctx context.Context, tx domain.Transaction) error
	SumIncomeBetween(ctx context.Context, userID ids.UserID, from, to time.Time) (int64, error)
	SumExpenseBetween(ctx context.Context, userID ids.UserID, from, to time.Time) (int64, error)
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserReader interface {
	Timezone(ctx context.Context, userID ids.UserID) (string, error)
}

type TransactionDTO struct {
	ID          ids.TransactionID
	Description string
	AmountCents int64
	Currency    string
	MonthTotal  int64
}

type RecordIncome struct {
	categories CategoryStore
	txs        TransactionStore
	events     EventLog
	transactor Transactor
	users      UserReader
	now        func() time.Time
}

func NewRecordIncome(
	categories CategoryStore,
	txs TransactionStore,
	events EventLog,
	transactor Transactor,
	users UserReader,
) *RecordIncome {
	return &RecordIncome{
		categories: categories,
		txs:        txs,
		events:     events,
		transactor: transactor,
		users:      users,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

type RecordIncomeInput struct {
	UserID      ids.UserID
	AmountCents int64
	Currency    string
	Description string
	Source      events.Source
}

func (uc *RecordIncome) Execute(ctx context.Context, in RecordIncomeInput) (TransactionDTO, error) {
	if in.UserID.IsZero() {
		return TransactionDTO{}, fmt.Errorf("user id is required")
	}
	if in.Currency == "" {
		in.Currency = "RUB"
	}

	now := uc.now()
	categoryID, err := uc.ensureIncomeCategory(ctx, in.UserID, now)
	if err != nil {
		return TransactionDTO{}, err
	}

	tx, err := domain.NewIncomeTransaction(
		in.UserID, categoryID, in.AmountCents, in.Currency, in.Description, now,
	)
	if err != nil {
		return TransactionDTO{}, err
	}

	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.txs.SaveTransaction(txCtx, tx); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        tx.UserID,
			AggregateType: "finance_transaction",
			AggregateID:   tx.ID.UUID(),
			EventType:     "IncomeRecorded",
			Payload: map[string]any{
				"amount_cents": tx.Money.AmountCents,
				"currency":     tx.Money.Currency,
				"description":  tx.Description,
			},
			Source:     in.Source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return TransactionDTO{}, fmt.Errorf("record income: %w", err)
	}

	monthTotal, err := uc.monthIncomeTotal(ctx, in.UserID, now)
	if err != nil {
		monthTotal = tx.Money.AmountCents
	}

	return TransactionDTO{
		ID:          tx.ID,
		Description: tx.Description,
		AmountCents: tx.Money.AmountCents,
		Currency:    tx.Money.Currency,
		MonthTotal:  monthTotal,
	}, nil
}

func (uc *RecordIncome) ensureIncomeCategory(ctx context.Context, userID ids.UserID, now time.Time) (ids.CategoryID, error) {
	cat, err := uc.categories.GetByName(ctx, userID, defaultIncomeCategory, domain.KindIncome)
	if err == nil {
		return cat.ID, nil
	}
	if !errors.Is(err, domain.ErrCategoryNotFound) {
		return ids.CategoryID{}, err
	}
	cat, err = domain.NewCategory(userID, defaultIncomeCategory, domain.KindIncome, now)
	if err != nil {
		return ids.CategoryID{}, err
	}
	if err := uc.categories.SaveCategory(ctx, cat); err != nil {
		return ids.CategoryID{}, err
	}
	return cat.ID, nil
}

func (uc *RecordIncome) monthIncomeTotal(ctx context.Context, userID ids.UserID, now time.Time) (int64, error) {
	start, end, err := monthBounds(ctx, uc.users, userID, now)
	if err != nil {
		return 0, err
	}
	return uc.txs.SumIncomeBetween(ctx, userID, start, end)
}
