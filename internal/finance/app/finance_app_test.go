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

type catStoreFake struct {
	byName map[string]domain.Category
	saved  []domain.Category
}

func (s *catStoreFake) GetByName(_ context.Context, _ ids.UserID, name string, kind domain.Kind) (domain.Category, error) {
	cat, ok := s.byName[name+"|"+string(kind)]
	if !ok {
		return domain.Category{}, domain.ErrCategoryNotFound
	}
	return cat, nil
}

func (s *catStoreFake) SaveCategory(_ context.Context, cat domain.Category) error {
	s.saved = append(s.saved, cat)
	if s.byName == nil {
		s.byName = map[string]domain.Category{}
	}
	s.byName[cat.Name+"|"+string(cat.Kind)] = cat
	return nil
}

type txStoreFake struct {
	saved   []domain.Transaction
	income  int64
	expense int64
	sumErr  error
}

func (s *txStoreFake) SaveTransaction(_ context.Context, tx domain.Transaction) error {
	s.saved = append(s.saved, tx)
	return nil
}

func (s *txStoreFake) SumIncomeBetween(context.Context, ids.UserID, time.Time, time.Time) (int64, error) {
	return s.income, s.sumErr
}

func (s *txStoreFake) SumExpenseBetween(context.Context, ids.UserID, time.Time, time.Time) (int64, error) {
	return s.expense, s.sumErr
}

type debtStoreFake struct {
	items []domain.Debt
}

func (s *debtStoreFake) SaveDebt(_ context.Context, debt domain.Debt) error {
	s.items = append(s.items, debt)
	return nil
}

func (s *debtStoreFake) ListOpen(_ context.Context, userID ids.UserID) ([]domain.Debt, error) {
	var out []domain.Debt
	for _, d := range s.items {
		if d.UserID == userID && d.Status == domain.DebtStatusOpen {
			out = append(out, d)
		}
	}
	return out, nil
}

func TestRecordIncomeAndExpense(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	cats := &catStoreFake{}
	txs := &txStoreFake{income: 15_000, expense: 7_500}
	ev := noopEvents{}
	tz := overviewTZFake{tz: "UTC"}

	incomeUC := app.NewRecordIncome(cats, txs, ev, noopTx{}, tz)
	income, err := incomeUC.Execute(context.Background(), app.RecordIncomeInput{
		UserID: userID, AmountCents: 5_000, Description: "аванс", Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if income.AmountCents != 5_000 || income.Currency != "RUB" || income.MonthTotal != 15_000 {
		t.Fatalf("income = %+v", income)
	}
	if len(cats.saved) != 1 || cats.saved[0].Name != "Доход" {
		t.Fatalf("income categories = %+v", cats.saved)
	}

	expenseUC := app.NewRecordExpense(cats, txs, ev, noopTx{}, tz)
	expense, err := expenseUC.Execute(context.Background(), app.RecordExpenseInput{
		UserID: userID, AmountCents: 2_000, Source: events.SourceTelegram,
	})
	if err != nil {
		t.Fatal(err)
	}
	if expense.Description != "Прочее" || expense.MonthTotal != 7_500 {
		t.Fatalf("expense = %+v", expense)
	}

	named, err := expenseUC.Execute(context.Background(), app.RecordExpenseInput{
		UserID: userID, AmountCents: 300, CategoryName: "Еда", Description: "обед",
		Currency: "RUB", Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if named.Description != "обед" {
		t.Fatalf("named = %+v", named)
	}
}

func TestRecordIncomeRequiresUser(t *testing.T) {
	t.Parallel()
	uc := app.NewRecordIncome(&catStoreFake{}, &txStoreFake{}, noopEvents{}, noopTx{}, overviewTZFake{tz: "UTC"})
	_, err := uc.Execute(context.Background(), app.RecordIncomeInput{AmountCents: 100, Description: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCashFlowSummary(t *testing.T) {
	t.Parallel()
	txs := &txStoreFake{income: 20_000, expense: 8_000}
	uc := app.NewCashFlowSummary(txs, overviewTZFake{tz: "Europe/Moscow"})
	dto, err := uc.Execute(context.Background(), ids.NewUserID())
	if err != nil {
		t.Fatal(err)
	}
	if dto.IncomeCents != 20_000 || dto.ExpenseCents != 8_000 || dto.NetCents != 12_000 {
		t.Fatalf("dto = %+v", dto)
	}
	if _, err := uc.Execute(context.Background(), ids.UserID{}); err == nil {
		t.Fatal("expected zero user error")
	}
}

func TestCreateAndListDebts(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	store := &debtStoreFake{}
	due := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	created, err := app.NewCreateDebt(store, noopEvents{}, noopTx{}).Execute(context.Background(), app.CreateDebtInput{
		UserID: userID, Creditor: "Иванов", AmountCents: 50_000, DueDate: &due, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Creditor != "Иванов" || created.RemainingCents != 50_000 {
		t.Fatalf("created = %+v", created)
	}

	list, err := app.NewListDebts(store).Execute(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("list = %+v", list)
	}

	if _, err := app.NewCreateDebt(store, noopEvents{}, noopTx{}).Execute(context.Background(), app.CreateDebtInput{
		Creditor: "x", AmountCents: 1, Source: events.SourceCLI,
	}); err == nil {
		t.Fatal("expected user id error")
	}
	if _, err := app.NewListDebts(store).Execute(context.Background(), ids.UserID{}); err == nil {
		t.Fatal("expected user id error")
	}
}

func TestPayDebtByIDAndValidation(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	debt, err := domain.NewDebt(userID, "банк", 200_000, nil, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	store := &debtPayStore{debt: debt}
	uc := app.NewPayDebt(store, noopEvents{}, noopTx{})

	dto, err := uc.Execute(context.Background(), app.PayDebtInput{
		UserID: userID, DebtID: debt.ID, AmountCents: 200_000, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dto.RemainingCents != 0 {
		t.Fatalf("remaining = %d", dto.RemainingCents)
	}

	if _, err := uc.Execute(context.Background(), app.PayDebtInput{
		UserID: userID, AmountCents: 1, Source: events.SourceCLI,
	}); err == nil {
		t.Fatal("expected debt id/creditor required")
	}
	if _, err := uc.Execute(context.Background(), app.PayDebtInput{
		UserID: userID, Creditor: "банк", AmountCents: 0, Source: events.SourceCLI,
	}); !errors.Is(err, domain.ErrInvalidAmount) {
		t.Fatalf("err = %v", err)
	}
	if _, err := uc.Execute(context.Background(), app.PayDebtInput{
		AmountCents: 1, Creditor: "банк", Source: events.SourceCLI,
	}); err == nil {
		t.Fatal("expected user id error")
	}
}
