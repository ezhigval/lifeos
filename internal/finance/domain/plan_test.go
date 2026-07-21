package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestAdvanceOccurrenceWeeklyCatchUp(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC)
	item, err := domain.NewPlannedCashflow(
		ids.NewUserID(), domain.PlanKindExpense, "Аренда", 10000,
		domain.PlanIntervalWeekly, time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC), now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if del := item.AdvanceOccurrence(now); del {
		t.Fatal("weekly should not delete")
	}
	want := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)
	if !item.NextDate.Equal(want) {
		t.Fatalf("next=%v want %v", item.NextDate, want)
	}
}

func TestAdvanceIfOverdueSkipsToday(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 21, 8, 0, 0, 0, time.UTC)
	item, err := domain.NewPlannedCashflow(
		ids.NewUserID(), domain.PlanKindIncome, "Зарплата", 100000,
		domain.PlanIntervalMonthly, time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC), now,
	)
	if err != nil {
		t.Fatal(err)
	}
	changed, del := item.AdvanceIfOverdue(now)
	if changed || del {
		t.Fatalf("today should stay visible: changed=%v del=%v", changed, del)
	}
}

func TestNewPlannedCashflowValidation(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC)
	uid := ids.NewUserID()
	if _, err := domain.NewPlannedCashflow(uid, domain.PlanKindIncome, "  ", 100, domain.PlanIntervalOnce, now, now); err != domain.ErrEmptyPlanTitle {
		t.Fatalf("empty title err=%v", err)
	}
	if _, err := domain.NewPlannedCashflow(uid, domain.PlanKind("x"), "t", 100, domain.PlanIntervalOnce, now, now); err != domain.ErrInvalidPlanKind {
		t.Fatalf("kind err=%v", err)
	}
	if _, err := domain.NewPlannedCashflow(uid, domain.PlanKindIncome, "t", 0, domain.PlanIntervalOnce, now, now); err != domain.ErrInvalidAmount {
		t.Fatalf("amount err=%v", err)
	}
	if _, err := domain.NewPlannedCashflow(uid, domain.PlanKindIncome, "t", 100, domain.PlanInterval("daily"), now, now); err != domain.ErrInvalidInterval {
		t.Fatalf("interval err=%v", err)
	}
	item, err := domain.NewPlannedCashflow(uid, domain.PlanKindIncome, "t", 100, "", now, now)
	if err != nil {
		t.Fatal(err)
	}
	if item.Interval != domain.PlanIntervalMonthly {
		t.Fatalf("default interval=%s", item.Interval)
	}
}

func TestAdvanceOccurrenceOnceAndMonthlyCatchUp(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC)
	once, err := domain.NewPlannedCashflow(ids.NewUserID(), domain.PlanKindExpense, "once", 10, domain.PlanIntervalOnce, now, now)
	if err != nil {
		t.Fatal(err)
	}
	if !once.AdvanceOccurrence(now) {
		t.Fatal("once should delete")
	}

	item, err := domain.NewPlannedCashflow(ids.NewUserID(), domain.PlanKindIncome, "m", 10, domain.PlanIntervalMonthly, time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC), now)
	if err != nil {
		t.Fatal(err)
	}
	changed, del := item.AdvanceIfOverdue(now)
	if !changed || del {
		t.Fatalf("changed=%v del=%v", changed, del)
	}
	want := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	if !item.NextDate.Equal(want) {
		t.Fatalf("next=%v want %v", item.NextDate, want)
	}
}
