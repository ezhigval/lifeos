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

func TestCompletePlanOccurrenceAdvancesMonthly(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	store := &memPlanStore{}
	fixed := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	item, err := domain.NewPlannedCashflow(user, domain.PlanKindIncome, "Зарплата", 100000, domain.PlanIntervalMonthly, time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC), fixed)
	if err != nil {
		t.Fatal(err)
	}
	_ = store.SavePlanned(context.Background(), item)

	uc := NewCompletePlanOccurrence(store, noopPlanEvents{}, noopPlanTx{}, nil, nil)
	uc.now = func() time.Time { return fixed }
	res, err := uc.Execute(context.Background(), user, item.ID, events.SourceHTTP)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted || res.Item == nil {
		t.Fatalf("want advanced item, got %+v", res)
	}
	if res.Item.NextDate != "2026-08-21" {
		t.Fatalf("next=%s", res.Item.NextDate)
	}
}

func TestAdvanceOverduePlansDeletesOnceBeforeToday(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	store := &memPlanStore{}
	fixed := time.Date(2026, 7, 21, 9, 0, 0, 0, time.UTC)
	past, err := domain.NewPlannedCashflow(user, domain.PlanKindExpense, "Старое", 2000, domain.PlanIntervalOnce, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC), fixed)
	if err != nil {
		t.Fatal(err)
	}
	_ = store.SavePlanned(context.Background(), past)

	today, err := domain.NewPlannedCashflow(user, domain.PlanKindIncome, "Сегодня", 3000, domain.PlanIntervalWeekly, time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC), fixed)
	if err != nil {
		t.Fatal(err)
	}
	_ = store.SavePlanned(context.Background(), today)

	uc := NewAdvanceOverduePlans(store, noopPlanEvents{}, noopPlanTx{})
	uc.now = func() time.Time { return fixed }
	res, err := uc.Execute(context.Background(), user, events.SourceScheduler)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted != 1 {
		t.Fatalf("deleted=%d want 1", res.Deleted)
	}
	if _, ok := store.items[today.ID]; !ok {
		t.Fatal("today item should remain")
	}
	if _, ok := store.items[past.ID]; ok {
		t.Fatal("past once item should be deleted")
	}
}
