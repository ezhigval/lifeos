package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

var (
	ErrEmptyPlanTitle  = errors.New("title is required")
	ErrInvalidPlanKind = errors.New("invalid planned cashflow kind")
	ErrInvalidInterval = errors.New("invalid planned cashflow interval")
	ErrPlanNotFound    = errors.New("planned cashflow not found")
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

// AdvanceOccurrence rolls next_date forward for one paid/passed occurrence.
// Returns shouldDelete=true for one-shot items that are done.
func (p *PlannedCashflow) AdvanceOccurrence(now time.Time) (shouldDelete bool) {
	nowDay := now.UTC().Truncate(24 * time.Hour)
	switch p.Interval {
	case PlanIntervalOnce:
		return true
	case PlanIntervalWeekly:
		p.NextDate = p.NextDate.UTC().Truncate(24*time.Hour).AddDate(0, 0, 7)
	case PlanIntervalMonthly:
		p.NextDate = p.NextDate.UTC().Truncate(24*time.Hour).AddDate(0, 1, 0)
	default:
		return false
	}
	// Catch up if several intervals were missed.
	for !p.NextDate.After(nowDay) {
		switch p.Interval {
		case PlanIntervalWeekly:
			p.NextDate = p.NextDate.AddDate(0, 0, 7)
		case PlanIntervalMonthly:
			p.NextDate = p.NextDate.AddDate(0, 1, 0)
		default:
			return false
		}
	}
	p.UpdatedAt = now.UTC()
	return false
}

// AdvanceIfOverdue rolls or deletes when next_date is strictly before today.
func (p *PlannedCashflow) AdvanceIfOverdue(now time.Time) (changed bool, shouldDelete bool) {
	nowDay := now.UTC().Truncate(24 * time.Hour)
	next := p.NextDate.UTC().Truncate(24 * time.Hour)
	if !next.Before(nowDay) {
		return false, false
	}
	shouldDelete = p.AdvanceOccurrence(now)
	return true, shouldDelete
}
