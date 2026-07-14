package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/app"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type overviewStoreFake struct {
	income     int64
	expense    int64
	categories []app.CategoryExpenseTotal
	from, to   time.Time
}

func (s *overviewStoreFake) SumIncomeBetween(_ context.Context, _ ids.UserID, from, to time.Time) (int64, error) {
	s.from, s.to = from, to
	return s.income, nil
}

func (s *overviewStoreFake) SumExpenseBetween(_ context.Context, _ ids.UserID, from, to time.Time) (int64, error) {
	s.from, s.to = from, to
	return s.expense, nil
}

func (s *overviewStoreFake) SumExpensesByCategoryBetween(_ context.Context, _ ids.UserID, from, to time.Time) ([]app.CategoryExpenseTotal, error) {
	s.from, s.to = from, to
	return s.categories, nil
}

type overviewTZFake struct {
	tz string
}

func (u overviewTZFake) Timezone(context.Context, ids.UserID) (string, error) {
	return u.tz, nil
}

func TestFinanceOverviewEmptyMonth(t *testing.T) {
	t.Parallel()
	store := &overviewStoreFake{categories: nil}
	uc := app.NewFinanceOverview(store, overviewTZFake{tz: "UTC"})

	dto, err := uc.Execute(context.Background(), app.FinanceOverviewInput{
		UserID: ids.NewUserID(),
		Period: "2026-07",
	})
	if err != nil {
		t.Fatal(err)
	}
	if dto.PeriodLabel != "июль 2026" {
		t.Fatalf("period_label = %q, want июль 2026", dto.PeriodLabel)
	}
	if dto.IncomeCents != 0 || dto.ExpenseCents != 0 || dto.NetCents != 0 {
		t.Fatalf("expected zeros, got income=%d expense=%d net=%d", dto.IncomeCents, dto.ExpenseCents, dto.NetCents)
	}
	if dto.Currency != "RUB" {
		t.Fatalf("currency = %q", dto.Currency)
	}
	if dto.Categories == nil || len(dto.Categories) != 0 {
		t.Fatalf("categories = %#v, want empty slice", dto.Categories)
	}
}

func TestFinanceOverviewCategoriesAndPercents(t *testing.T) {
	t.Parallel()
	store := &overviewStoreFake{
		income:  10_000,
		expense: 5_000,
		categories: []app.CategoryExpenseTotal{
			{Name: "Еда", AmountCents: 3_000},
			{Name: "Транспорт", AmountCents: 2_000},
		},
	}
	uc := app.NewFinanceOverview(store, overviewTZFake{tz: "Europe/Moscow"})

	dto, err := uc.Execute(context.Background(), app.FinanceOverviewInput{
		UserID: ids.NewUserID(),
		Period: "2026-07",
	})
	if err != nil {
		t.Fatal(err)
	}
	if dto.NetCents != 5_000 {
		t.Fatalf("net = %d, want 5000", dto.NetCents)
	}
	if len(dto.Categories) != 2 {
		t.Fatalf("categories len = %d", len(dto.Categories))
	}
	if dto.Categories[0].Name != "Еда" || dto.Categories[0].Percent != 60 {
		t.Fatalf("first category = %+v, want Еда 60%%", dto.Categories[0])
	}
	if dto.Categories[1].Name != "Транспорт" || dto.Categories[1].Percent != 40 {
		t.Fatalf("second category = %+v, want Транспорт 40%%", dto.Categories[1])
	}

	// July 2026 in Europe/Moscow starts at 2026-06-30 21:00 UTC
	wantStart := time.Date(2026, 6, 30, 21, 0, 0, 0, time.UTC)
	if !store.from.Equal(wantStart) {
		t.Fatalf("month start = %v, want %v", store.from, wantStart)
	}
	wantEnd := time.Date(2026, 7, 31, 21, 0, 0, 0, time.UTC)
	if !store.to.Equal(wantEnd) {
		t.Fatalf("month end = %v, want %v", store.to, wantEnd)
	}
}

func TestFinanceOverviewInvalidPeriod(t *testing.T) {
	t.Parallel()
	uc := app.NewFinanceOverview(&overviewStoreFake{}, overviewTZFake{tz: "UTC"})
	userID := ids.NewUserID()

	for _, period := range []string{"", "2026-7", "2026-13", "26-07", "2026/07", "july-2026"} {
		_, err := uc.Execute(context.Background(), app.FinanceOverviewInput{
			UserID: userID,
			Period: period,
		})
		if err != app.ErrInvalidPeriod {
			t.Fatalf("period %q: err = %v, want ErrInvalidPeriod", period, err)
		}
	}
}
