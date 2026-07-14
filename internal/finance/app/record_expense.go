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

const defaultExpenseCategory = "Прочее"

type RecordExpense struct {
	categories CategoryStore
	txs        TransactionStore
	events     EventLog
	transactor Transactor
	users      UserReader
	now        func() time.Time
}

func NewRecordExpense(
	categories CategoryStore,
	txs TransactionStore,
	events EventLog,
	transactor Transactor,
	users UserReader,
) *RecordExpense {
	return &RecordExpense{
		categories: categories,
		txs:        txs,
		events:     events,
		transactor: transactor,
		users:      users,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

type RecordExpenseInput struct {
	UserID       ids.UserID
	AmountCents  int64
	Currency     string
	CategoryName string
	Description  string
	Source       events.Source
}

func (uc *RecordExpense) Execute(ctx context.Context, in RecordExpenseInput) (TransactionDTO, error) {
	if in.UserID.IsZero() {
		return TransactionDTO{}, fmt.Errorf("user id is required")
	}
	if in.Currency == "" {
		in.Currency = "RUB"
	}
	if in.CategoryName == "" {
		in.CategoryName = defaultExpenseCategory
	}
	if in.Description == "" {
		in.Description = in.CategoryName
	}

	now := uc.now()
	categoryID, err := uc.ensureExpenseCategory(ctx, in.UserID, in.CategoryName, now)
	if err != nil {
		return TransactionDTO{}, err
	}

	tx, err := domain.NewExpenseTransaction(
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
			EventType:     "ExpenseRecorded",
			Payload: map[string]any{
				"amount_cents": tx.Money.AmountCents,
				"currency":     tx.Money.Currency,
				"description":  tx.Description,
				"category":     in.CategoryName,
			},
			Source:     in.Source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return TransactionDTO{}, fmt.Errorf("record expense: %w", err)
	}

	monthTotal, err := uc.monthExpenseTotal(ctx, in.UserID, now)
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

func (uc *RecordExpense) ensureExpenseCategory(ctx context.Context, userID ids.UserID, name string, now time.Time) (ids.CategoryID, error) {
	cat, err := uc.categories.GetByName(ctx, userID, name, domain.KindExpense)
	if err == nil {
		return cat.ID, nil
	}
	if !errors.Is(err, domain.ErrCategoryNotFound) {
		return ids.CategoryID{}, err
	}
	cat, err = domain.NewCategory(userID, name, domain.KindExpense, now)
	if err != nil {
		return ids.CategoryID{}, err
	}
	if err := uc.categories.SaveCategory(ctx, cat); err != nil {
		return ids.CategoryID{}, err
	}
	return cat.ID, nil
}

func (uc *RecordExpense) monthExpenseTotal(ctx context.Context, userID ids.UserID, now time.Time) (int64, error) {
	start, end, err := monthBounds(ctx, uc.users, userID, now)
	if err != nil {
		return 0, err
	}
	return uc.txs.SumExpenseBetween(ctx, userID, start, end)
}
