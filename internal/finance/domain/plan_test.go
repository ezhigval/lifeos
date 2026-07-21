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
