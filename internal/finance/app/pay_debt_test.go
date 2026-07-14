package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/app"
	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type debtPayStore struct {
	debt domain.Debt
}

func (s *debtPayStore) GetByID(_ context.Context, _ ids.UserID, _ ids.DebtID) (domain.Debt, error) {
	return s.debt, nil
}

func (s *debtPayStore) FindOpenByCreditor(_ context.Context, _ ids.UserID, creditor string) (domain.Debt, error) {
	if creditor == "банку" {
		return s.debt, nil
	}
	return domain.Debt{}, domain.ErrDebtNotFound
}

func (s *debtPayStore) UpdateDebt(_ context.Context, debt domain.Debt) error {
	s.debt = debt
	return nil
}

type noopEvents struct{}

func (noopEvents) Append(context.Context, events.Record) error { return nil }

type noopTx struct{}

func (noopTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestPayDebtByCreditor(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	debt, err := domain.NewDebt(userID, "банку", 1_000_000, nil, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	store := &debtPayStore{debt: debt}
	uc := app.NewPayDebt(store, noopEvents{}, noopTx{})

	dto, err := uc.Execute(context.Background(), app.PayDebtInput{
		UserID: userID, Creditor: "банку", AmountCents: 100_000,
		Source: events.SourceTelegram,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dto.PaidCents != 100_000 {
		t.Fatalf("paid = %d, want 100000", dto.PaidCents)
	}
	if dto.RemainingCents != 900_000 {
		t.Fatalf("remaining = %d, want 900000", dto.RemainingCents)
	}
}

func TestPayDebtOverpayment(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	debt, _ := domain.NewDebt(userID, "банку", 50_000, nil, time.Now().UTC())
	store := &debtPayStore{debt: debt}
	uc := app.NewPayDebt(store, noopEvents{}, noopTx{})

	_, err := uc.Execute(context.Background(), app.PayDebtInput{
		UserID: userID, Creditor: "банку", AmountCents: 100_000,
		Source: events.SourceTelegram,
	})
	if !errors.Is(err, domain.ErrOverpayment) {
		t.Fatalf("err = %v, want ErrOverpayment", err)
	}
}
