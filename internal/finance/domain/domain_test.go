package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestNewMoneyDefaultsCurrencyAndFormatKopecks(t *testing.T) {
	t.Parallel()
	m, err := domain.NewMoney(150, "")
	if err != nil {
		t.Fatal(err)
	}
	if m.Currency != "RUB" {
		t.Fatalf("currency = %q", m.Currency)
	}
	if got := domain.FormatMoney(m); got != "1,50 ₽" {
		t.Fatalf("FormatMoney = %q", got)
	}
}

func TestKindValid(t *testing.T) {
	t.Parallel()
	if !domain.KindIncome.Valid() || !domain.KindExpense.Valid() {
		t.Fatal("income/expense should be valid")
	}
	if domain.Kind("transfer").Valid() {
		t.Fatal("unknown kind should be invalid")
	}
}

func TestNewCategory(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	cat, err := domain.NewCategory(ids.NewUserID(), "Еда", domain.KindExpense, now)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Name != "Еда" || cat.Kind != domain.KindExpense || cat.ID.IsZero() {
		t.Fatalf("cat = %+v", cat)
	}
	if !cat.CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt = %v", cat.CreatedAt)
	}

	if _, err := domain.NewCategory(ids.NewUserID(), "", domain.KindIncome, now); !errors.Is(err, domain.ErrEmptyCategoryName) {
		t.Fatalf("empty name err = %v", err)
	}
	if _, err := domain.NewCategory(ids.NewUserID(), "X", domain.Kind("bad"), now); !errors.Is(err, domain.ErrInvalidKind) {
		t.Fatalf("bad kind err = %v", err)
	}
}

func TestNewTransactions(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	catID := ids.NewCategoryID()
	at := time.Date(2026, 7, 1, 10, 0, 0, 0, time.FixedZone("MSK", 3*3600))

	income, err := domain.NewIncomeTransaction(userID, catID, 10_000, "RUB", "зарплата", at)
	if err != nil {
		t.Fatal(err)
	}
	if income.Kind != domain.KindIncome || income.Description != "зарплата" {
		t.Fatalf("income = %+v", income)
	}
	if !income.OccurredAt.Equal(at.UTC()) {
		t.Fatalf("OccurredAt = %v want UTC", income.OccurredAt)
	}

	expense, err := domain.NewExpenseTransaction(userID, catID, 250, "USD", "кофе", at)
	if err != nil {
		t.Fatal(err)
	}
	if expense.Kind != domain.KindExpense || expense.Money.Currency != "USD" {
		t.Fatalf("expense = %+v", expense)
	}

	if _, err := domain.NewIncomeTransaction(userID, catID, 100, "RUB", "", at); !errors.Is(err, domain.ErrEmptyDescription) {
		t.Fatalf("empty description err = %v", err)
	}
	if _, err := domain.NewExpenseTransaction(userID, catID, 0, "RUB", "x", at); !errors.Is(err, domain.ErrInvalidAmount) {
		t.Fatalf("zero amount err = %v", err)
	}
}

func TestDebtLifecycle(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	now := time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC)
	due := now.Add(48 * time.Hour)

	if !domain.DebtStatusOpen.Valid() || !domain.DebtStatusPaid.Valid() || !domain.DebtStatusCancelled.Valid() {
		t.Fatal("known statuses should be valid")
	}
	if domain.DebtStatus("frozen").Valid() {
		t.Fatal("unknown status should be invalid")
	}

	if _, err := domain.NewDebt(userID, "", 100, nil, now); !errors.Is(err, domain.ErrEmptyCreditor) {
		t.Fatalf("empty creditor err = %v", err)
	}
	if _, err := domain.NewDebt(userID, "банк", 0, nil, now); !errors.Is(err, domain.ErrInvalidAmount) {
		t.Fatalf("zero amount err = %v", err)
	}

	debt, err := domain.NewDebt(userID, "банк", 1_000, &due, now)
	if err != nil {
		t.Fatal(err)
	}
	if debt.Status != domain.DebtStatusOpen || debt.RemainingCents() != 1_000 || debt.PaidCents != 0 {
		t.Fatalf("debt = %+v", debt)
	}
	if debt.DueDate == nil || !debt.DueDate.Equal(due) {
		t.Fatalf("due = %v", debt.DueDate)
	}

	if err := debt.Pay(0); !errors.Is(err, domain.ErrInvalidAmount) {
		t.Fatalf("pay zero err = %v", err)
	}
	if err := debt.Pay(1_500); !errors.Is(err, domain.ErrOverpayment) {
		t.Fatalf("overpay err = %v", err)
	}
	if err := debt.Pay(400); err != nil {
		t.Fatal(err)
	}
	if debt.RemainingCents() != 600 || debt.Status != domain.DebtStatusOpen {
		t.Fatalf("after partial pay = %+v", debt)
	}
	if err := debt.Pay(600); err != nil {
		t.Fatal(err)
	}
	if debt.Status != domain.DebtStatusPaid || debt.RemainingCents() != 0 {
		t.Fatalf("after full pay = %+v", debt)
	}
	if err := debt.Pay(1); !errors.Is(err, domain.ErrDebtNotOpen) {
		t.Fatalf("pay closed err = %v", err)
	}
}

func TestDebtAdvanceInstallment(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	due := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	debt, err := domain.NewDebt(ids.NewUserID(), "банк", 10_000, &due, now)
	if err != nil {
		t.Fatal(err)
	}
	debt.AdvanceInstallment(now) // no-op without installment
	if debt.NextPaymentDate != nil {
		t.Fatalf("noop should keep next nil, got %v", debt.NextPaymentDate)
	}
	debt.InstallmentCents = 1000
	debt.InstallmentInterval = "weekly"
	debt.NextPaymentDate = &due
	debt.AdvanceInstallment(now)
	want := due.AddDate(0, 0, 7)
	if debt.NextPaymentDate == nil || !debt.NextPaymentDate.Equal(want) {
		t.Fatalf("weekly next=%v want %v", debt.NextPaymentDate, want)
	}
	debt.InstallmentInterval = "monthly"
	base := *debt.NextPaymentDate
	debt.AdvanceInstallment(now)
	wantM := base.AddDate(0, 1, 0)
	if !debt.NextPaymentDate.Equal(wantM) {
		t.Fatalf("monthly next=%v want %v", debt.NextPaymentDate, wantM)
	}
}
