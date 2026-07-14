package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type CashFlowSummary struct {
	txs   TransactionStore
	users UserReader
	now   func() time.Time
}

func NewCashFlowSummary(txs TransactionStore, users UserReader) *CashFlowSummary {
	return &CashFlowSummary{
		txs:   txs,
		users: users,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

type CashFlowDTO struct {
	IncomeCents  int64
	ExpenseCents int64
	NetCents     int64
	Currency     string
}

func (uc *CashFlowSummary) Execute(ctx context.Context, userID ids.UserID) (CashFlowDTO, error) {
	if userID.IsZero() {
		return CashFlowDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	start, end, err := monthBounds(ctx, uc.users, userID, now)
	if err != nil {
		return CashFlowDTO{}, err
	}
	income, err := uc.txs.SumIncomeBetween(ctx, userID, start, end)
	if err != nil {
		return CashFlowDTO{}, err
	}
	expense, err := uc.txs.SumExpenseBetween(ctx, userID, start, end)
	if err != nil {
		return CashFlowDTO{}, err
	}
	return CashFlowDTO{
		IncomeCents:  income,
		ExpenseCents: expense,
		NetCents:     income - expense,
		Currency:     "RUB",
	}, nil
}
