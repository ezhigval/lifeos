package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyPlanTitle   = errors.New("title is required")
	ErrInvalidPlanKind  = errors.New("invalid planned cashflow kind")
	ErrInvalidInterval  = errors.New("invalid planned cashflow interval")
	ErrPlanNotFound     = errors.New("planned cashflow not found")
)

type PlanKind string

const (
	PlanKindIncome  PlanKind = "income"
	PlanKindExpense PlanKind = "expense"
)

func (k PlanKind) Valid() bool {
	return k == PlanKindIncome || k == PlanKindExpense
}

type PlanInterval string

const (
	PlanIntervalOnce    PlanInterval = "once"
	PlanIntervalWeekly  PlanInterval = "weekly"
	PlanIntervalMonthly PlanInterval = "monthly"
)

func (i PlanInterval) Valid() bool {
	return i == PlanIntervalOnce || i == PlanIntervalWeekly || i == PlanIntervalMonthly
}

type PlannedCashflow struct {
	ID          ids.PlannedCashflowID
	UserID      ids.UserID
	Kind        PlanKind
	Title       string
	AmountCents int64
	Interval    PlanInterval
	NextDate    time.Time
	DebtID      *ids.DebtID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewPlannedCashflow(
	userID ids.UserID,
	kind PlanKind,
	title string,
	amountCents int64,
	interval PlanInterval,
	nextDate time.Time,
	now time.Time,
) (PlannedCashflow, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return PlannedCashflow{}, ErrEmptyPlanTitle
	}
	if !kind.Valid() {
		return PlannedCashflow{}, ErrInvalidPlanKind
	}
	if amountCents <= 0 {
		return PlannedCashflow{}, ErrInvalidAmount
	}
	if interval == "" {
		interval = PlanIntervalMonthly
	}
	if !interval.Valid() {
		return PlannedCashflow{}, ErrInvalidInterval
	}
	now = now.UTC()
	return PlannedCashflow{
		ID:          ids.NewPlannedCashflowID(),
		UserID:      userID,
		Kind:        kind,
		Title:       title,
		AmountCents: amountCents,
		Interval:    interval,
		NextDate:    nextDate.UTC().Truncate(24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
