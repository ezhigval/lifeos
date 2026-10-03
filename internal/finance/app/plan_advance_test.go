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

	uc := NewAdvanceOverduePlans(store, noopPlanEvents{}, noopPlanTx{}, nil, nil)
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

type memCatStore struct {
	byName map[string]domain.Category
}

func (s *memCatStore) GetByName(_ context.Context, _ ids.UserID, name string, kind domain.Kind) (domain.Category, error) {
	cat, ok := s.byName[name+"|"+string(kind)]
	if !ok {
		return domain.Category{}, domain.ErrCategoryNotFound
	}
	return cat, nil
}
func (s *memCatStore) SaveCategory(_ context.Context, cat domain.Category) error {
	if s.byName == nil {
		s.byName = map[string]domain.Category{}
	}
	s.byName[cat.Name+"|"+string(cat.Kind)] = cat
	return nil
}

type memTxStore struct {
	saved []domain.Transaction
}

func (s *memTxStore) SaveTransaction(_ context.Context, tx domain.Transaction) error {
	s.saved = append(s.saved, tx)
	return nil
}
func (s *memTxStore) SumIncomeBetween(context.Context, ids.UserID, time.Time, time.Time) (int64, error) {
	return 0, nil
}
func (s *memTxStore) SumExpenseBetween(context.Context, ids.UserID, time.Time, time.Time) (int64, error) {
	return 0, nil
}

type planTZ struct{ tz string }

func (p planTZ) Timezone(context.Context, ids.UserID) (string, error) { return p.tz, nil }

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

func TestCompletePlanOccurrencePostsExpenseAndDeletesOnce(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	store := &memPlanStore{}
	txs := &memTxStore{}
	cats := &memCatStore{}
	fixed := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	item, err := domain.NewPlannedCashflow(user, domain.PlanKindExpense, "Разовое", 2500, domain.PlanIntervalOnce, time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC), fixed)
	if err != nil {
		t.Fatal(err)
	}
	_ = store.SavePlanned(context.Background(), item)

	expense := NewRecordExpense(cats, txs, noopPlanEvents{}, noopPlanTx{}, planTZ{tz: "UTC"})
	expense.now = func() time.Time { return fixed }
	uc := NewCompletePlanOccurrence(store, noopPlanEvents{}, noopPlanTx{}, nil, expense)
	uc.now = func() time.Time { return fixed }
	res, err := uc.Execute(context.Background(), user, item.ID, events.SourceHTTP)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Deleted || !res.Posted || res.PostedKind != "expense" || res.PostedCents != 2500 {
		t.Fatalf("res=%+v", res)
	}
	if len(txs.saved) != 1 || txs.saved[0].Kind != domain.KindExpense {
		t.Fatalf("txs=%+v", txs.saved)
	}
}

func TestCompletePlanOccurrencePostsIncome(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	store := &memPlanStore{}
	txs := &memTxStore{}
	cats := &memCatStore{}
	fixed := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	item, err := domain.NewPlannedCashflow(user, domain.PlanKindIncome, "Фриланс", 80000, domain.PlanIntervalMonthly, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), fixed)
	if err != nil {
		t.Fatal(err)
	}
	_ = store.SavePlanned(context.Background(), item)

	income := NewRecordIncome(cats, txs, noopPlanEvents{}, noopPlanTx{}, planTZ{tz: "UTC"})
	income.now = func() time.Time { return fixed }
	uc := NewCompletePlanOccurrence(store, noopPlanEvents{}, noopPlanTx{}, income, nil)
	uc.now = func() time.Time { return fixed }
	res, err := uc.Execute(context.Background(), user, item.ID, events.SourceHTTP)
	if err != nil {
		t.Fatal(err)
	}
	if res.Deleted || !res.Posted || res.PostedKind != "income" {
		t.Fatalf("res=%+v", res)
	}
	if len(txs.saved) != 1 || txs.saved[0].Kind != domain.KindIncome {
		t.Fatalf("txs=%+v", txs.saved)
	}
}

func TestAdvanceOverduePlansPostsLedger(t *testing.T) {
	t.Parallel()
	user := ids.NewUserID()
	store := &memPlanStore{}
	txs := &memTxStore{}
	cats := &memCatStore{}
	fixed := time.Date(2026, 7, 21, 9, 0, 0, 0, time.UTC)
	past, err := domain.NewPlannedCashflow(user, domain.PlanKindIncome, "Старый доход", 1000, domain.PlanIntervalWeekly, time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC), fixed)
	if err != nil {
		t.Fatal(err)
	}
	_ = store.SavePlanned(context.Background(), past)

	income := NewRecordIncome(cats, txs, noopPlanEvents{}, noopPlanTx{}, planTZ{tz: "UTC"})
	income.now = func() time.Time { return fixed }
	uc := NewAdvanceOverduePlans(store, noopPlanEvents{}, noopPlanTx{}, income, nil)
	uc.now = func() time.Time { return fixed }
	res, err := uc.Execute(context.Background(), user, events.SourceScheduler)
	if err != nil {
		t.Fatal(err)
	}
	if res.Advanced != 1 || res.Posted != 1 {
		t.Fatalf("res=%+v", res)
	}
	if len(txs.saved) != 1 {
		t.Fatalf("txs=%d", len(txs.saved))
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
