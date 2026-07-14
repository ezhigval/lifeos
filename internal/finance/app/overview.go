package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
)

var ErrInvalidPeriod = errors.New("invalid period")

type CategoryExpenseTotal struct {
	Name        string
	AmountCents int64
}

type OverviewStore interface {
	SumIncomeBetween(ctx context.Context, userID ids.UserID, from, to time.Time) (int64, error)
	SumExpenseBetween(ctx context.Context, userID ids.UserID, from, to time.Time) (int64, error)
	SumExpensesByCategoryBetween(ctx context.Context, userID ids.UserID, from, to time.Time) ([]CategoryExpenseTotal, error)
}

type FinanceOverview struct {
	txs   OverviewStore
	users UserReader
}

func NewFinanceOverview(txs OverviewStore, users UserReader) *FinanceOverview {
	return &FinanceOverview{txs: txs, users: users}
}

type FinanceOverviewInput struct {
	UserID ids.UserID
	Period string // YYYY-MM
}

type FinanceCategoryDTO struct {
	Name        string
	AmountCents int64
	Percent     float64
	ColorHint   *string
}

type FinanceOverviewDTO struct {
	PeriodLabel  string
	IncomeCents  int64
	ExpenseCents int64
	NetCents     int64
	Currency     string
	Categories   []FinanceCategoryDTO
}

func (uc *FinanceOverview) Execute(ctx context.Context, in FinanceOverviewInput) (FinanceOverviewDTO, error) {
	if in.UserID.IsZero() {
		return FinanceOverviewDTO{}, fmt.Errorf("user id is required")
	}
	year, month, err := parsePeriodKey(in.Period)
	if err != nil {
		return FinanceOverviewDTO{}, err
	}

	tz, err := uc.users.Timezone(ctx, in.UserID)
	if err != nil {
		return FinanceOverviewDTO{}, err
	}
	start, end, err := timeutil.MonthBoundsForPeriod(year, month, tz)
	if err != nil {
		return FinanceOverviewDTO{}, fmt.Errorf("%w: %v", ErrInvalidPeriod, err)
	}

	income, err := uc.txs.SumIncomeBetween(ctx, in.UserID, start, end)
	if err != nil {
		return FinanceOverviewDTO{}, err
	}
	expense, err := uc.txs.SumExpenseBetween(ctx, in.UserID, start, end)
	if err != nil {
		return FinanceOverviewDTO{}, err
	}

	rows, err := uc.txs.SumExpensesByCategoryBetween(ctx, in.UserID, start, end)
	if err != nil {
		return FinanceOverviewDTO{}, err
	}

	categories := make([]FinanceCategoryDTO, 0, len(rows))
	for _, row := range rows {
		percent := 0.0
		if expense > 0 {
			percent = float64(row.AmountCents) / float64(expense) * 100
		}
		categories = append(categories, FinanceCategoryDTO{
			Name:        row.Name,
			AmountCents: row.AmountCents,
			Percent:     percent,
		})
	}

	return FinanceOverviewDTO{
		PeriodLabel:  periodFullLabelRU(year, month),
		IncomeCents:  income,
		ExpenseCents: expense,
		NetCents:     income - expense,
		Currency:     "RUB",
		Categories:   categories,
	}, nil
}

func parsePeriodKey(period string) (int, time.Month, error) {
	if len(period) != 7 || period[4] != '-' {
		return 0, 0, ErrInvalidPeriod
	}
	var year, month int
	if _, err := fmt.Sscanf(period, "%d-%d", &year, &month); err != nil {
		return 0, 0, ErrInvalidPeriod
	}
	// Require zero-padded month (YYYY-MM), reject e.g. 2026-7
	if fmt.Sprintf("%04d-%02d", year, month) != period {
		return 0, 0, ErrInvalidPeriod
	}
	if month < 1 || month > 12 {
		return 0, 0, ErrInvalidPeriod
	}
	return year, time.Month(month), nil
}

var monthsRU = [...]string{
	"", "январь", "февраль", "март", "апрель", "май", "июнь",
	"июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь",
}

func periodFullLabelRU(year int, month time.Month) string {
	return fmt.Sprintf("%s %d", monthsRU[month], year)
}
