package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

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
	if !domain.PlanKindIncome.Valid() || !domain.PlanIntervalWeekly.Valid() {
		t.Fatal("valid helpers")
	}
}
