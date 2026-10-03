package app

import (
	"context"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type memPlanStore struct {
	items map[ids.PlannedCashflowID]domain.PlannedCashflow
}

func (s *memPlanStore) SavePlanned(_ context.Context, item domain.PlannedCashflow) error {
	if s.items == nil {
		s.items = map[ids.PlannedCashflowID]domain.PlannedCashflow{}
	}
	s.items[item.ID] = item
	return nil
}
func (s *memPlanStore) GetPlanned(_ context.Context, userID ids.UserID, id ids.PlannedCashflowID) (domain.PlannedCashflow, error) {
	p, ok := s.items[id]
	if !ok || p.UserID != userID {
		return domain.PlannedCashflow{}, domain.ErrPlanNotFound
	}
	return p, nil
}
func (s *memPlanStore) UpdatePlannedNextDate(_ context.Context, item domain.PlannedCashflow) error {
	s.items[item.ID] = item
	return nil
}
func (s *memPlanStore) ListPlanned(_ context.Context, userID ids.UserID) ([]domain.PlannedCashflow, error) {
	var out []domain.PlannedCashflow
	for _, p := range s.items {
		if p.UserID == userID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (s *memPlanStore) DeletePlanned(_ context.Context, userID ids.UserID, id ids.PlannedCashflowID) error {
	p, ok := s.items[id]
	if !ok || p.UserID != userID {
		return domain.ErrPlanNotFound
	}
	delete(s.items, id)
	return nil
}

type noopPlanEvents struct{}

func (noopPlanEvents) Append(context.Context, events.Record) error { return nil }

type noopPlanTx struct{}

func (noopPlanTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type memDebtStore struct{ items []domain.Debt }

func (s *memDebtStore) SaveDebt(_ context.Context, d domain.Debt) error {
	s.items = append(s.items, d)
	return nil
}
func (s *memDebtStore) ListOpen(_ context.Context, userID ids.UserID) ([]domain.Debt, error) {
	var out []domain.Debt
	for _, d := range s.items {
		if d.UserID == userID && d.Status == domain.DebtStatusOpen {
			out = append(out, d)
		}
	}
	return out, nil
}

func TestCreateListDeletePlannedCashflow(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	store := &memPlanStore{}
	fixed := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)

	create := NewCreatePlannedCashflow(store, noopPlanEvents{}, noopPlanTx{})
	create.now = func() time.Time { return fixed }
	item, err := create.Execute(context.Background(), CreatePlannedCashflowInput{
		UserID: user, Kind: "expense", Title: "Аренда", AmountCents: 50000,
		Interval: "monthly", NextDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Source: events.SourceHTTP,
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "Аренда" || item.Kind != "expense" || item.NextDate != "2026-08-01" {
		t.Fatalf("item=%+v", item)
	}

	due := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	debt, err := domain.NewDebt(user, "Банк", 100000, &due, fixed)
	if err != nil {
		t.Fatal(err)
	}
	debt.InstallmentCents = 10000
	debt.InstallmentInterval = "monthly"
	debt.NextPaymentDate = &due
	debts := &memDebtStore{items: []domain.Debt{debt}}

	listed, err := NewListFinancePlan(store, debts).Execute(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}
	if listed.PlannedExpense != 60000 || len(listed.Items) != 2 {
		t.Fatalf("plan=%+v", listed)
	}

	planID, err := ids.ParsePlannedCashflowID(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewDeletePlannedCashflow(store).Execute(context.Background(), user, planID); err != nil {
		t.Fatal(err)
	}
	if len(store.items) != 0 {
		t.Fatalf("store still has %d", len(store.items))
	}
}

func TestCreatePlannedCashflowValidation(t *testing.T) {
	t.Parallel()
	uc := NewCreatePlannedCashflow(&memPlanStore{}, noopPlanEvents{}, noopPlanTx{})
	if _, err := uc.Execute(context.Background(), CreatePlannedCashflowInput{}); err == nil {
		t.Fatal("expected error")
	}
	if err := NewDeletePlannedCashflow(&memPlanStore{}).Execute(context.Background(), ids.UserID{}, ids.PlannedCashflowID{}); err == nil {
		t.Fatal("expected error")
	}
	if _, err := NewListFinancePlan(&memPlanStore{}, &memDebtStore{}).Execute(context.Background(), ids.UserID{}); err == nil {
		t.Fatal("expected error")
	}
}
