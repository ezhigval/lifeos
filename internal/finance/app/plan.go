package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type PlannedCashflowStore interface {
	SavePlanned(ctx context.Context, item domain.PlannedCashflow) error
	ListPlanned(ctx context.Context, userID ids.UserID) ([]domain.PlannedCashflow, error)
	DeletePlanned(ctx context.Context, userID ids.UserID, id ids.PlannedCashflowID) error
}

type PlanItemDTO struct {
	ID          string
	Kind        string // income | expense
	Title       string
	AmountCents int64
	Currency    string
	Interval    string
	NextDate    string // YYYY-MM-DD
	Source      string // plan | debt_installment
	DebtID      *string
}

type FinancePlanDTO struct {
	Items           []PlanItemDTO
	PlannedIncome   int64
	PlannedExpense  int64
	Currency        string
}

type CreatePlannedCashflow struct {
	store      PlannedCashflowStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreatePlannedCashflow(store PlannedCashflowStore, events EventLog, transactor Transactor) *CreatePlannedCashflow {
	return &CreatePlannedCashflow{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreatePlannedCashflowInput struct {
	UserID      ids.UserID
	Kind        string
	Title       string
	AmountCents int64
	Interval    string
	NextDate    time.Time
	Source      events.Source
}

func (uc *CreatePlannedCashflow) Execute(ctx context.Context, in CreatePlannedCashflowInput) (PlanItemDTO, error) {
	if in.UserID.IsZero() {
		return PlanItemDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	item, err := domain.NewPlannedCashflow(
		in.UserID,
		domain.PlanKind(strings.TrimSpace(in.Kind)),
		in.Title,
		in.AmountCents,
		domain.PlanInterval(strings.TrimSpace(in.Interval)),
		in.NextDate,
		now,
	)
	if err != nil {
		return PlanItemDTO{}, err
	}
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.SavePlanned(txCtx, item); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        item.UserID,
			AggregateType: "planned_cashflow",
			AggregateID:   item.ID.UUID(),
			EventType:     "PlannedCashflowCreated",
			Payload: map[string]any{
				"kind":         item.Kind,
				"title":        item.Title,
				"amount_cents": item.AmountCents,
				"interval":     item.Interval,
			},
			Source:     in.Source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return PlanItemDTO{}, fmt.Errorf("create planned cashflow: %w", err)
	}
	return plannedToDTO(item), nil
}

type ListFinancePlan struct {
	plans PlannedCashflowStore
	debts DebtStore
}

func NewListFinancePlan(plans PlannedCashflowStore, debts DebtStore) *ListFinancePlan {
	return &ListFinancePlan{plans: plans, debts: debts}
}

func (uc *ListFinancePlan) Execute(ctx context.Context, userID ids.UserID) (FinancePlanDTO, error) {
	if userID.IsZero() {
		return FinancePlanDTO{}, fmt.Errorf("user id is required")
	}
	planned, err := uc.plans.ListPlanned(ctx, userID)
	if err != nil {
		return FinancePlanDTO{}, err
	}
	openDebts, err := uc.debts.ListOpen(ctx, userID)
	if err != nil {
		return FinancePlanDTO{}, err
	}

	items := make([]PlanItemDTO, 0, len(planned)+len(openDebts))
	var income, expense int64
	for _, p := range planned {
		dto := plannedToDTO(p)
		items = append(items, dto)
		if dto.Kind == string(domain.PlanKindIncome) {
			income += dto.AmountCents
		} else {
			expense += dto.AmountCents
		}
	}
	for _, d := range openDebts {
		if d.InstallmentCents <= 0 || d.InstallmentInterval == "none" || d.NextPaymentDate == nil {
			continue
		}
		id := d.ID.String()
		dto := PlanItemDTO{
			ID:          "debt-" + id,
			Kind:        string(domain.PlanKindExpense),
			Title:       "Платёж · " + d.Creditor,
			AmountCents: d.InstallmentCents,
			Currency:    "RUB",
			Interval:    d.InstallmentInterval,
			NextDate:    d.NextPaymentDate.Format("2006-01-02"),
			Source:      "debt_installment",
			DebtID:      &id,
		}
		items = append(items, dto)
		expense += dto.AmountCents
	}
	return FinancePlanDTO{
		Items:          items,
		PlannedIncome:  income,
		PlannedExpense: expense,
		Currency:       "RUB",
	}, nil
}

type DeletePlannedCashflow struct {
	store PlannedCashflowStore
}

func NewDeletePlannedCashflow(store PlannedCashflowStore) *DeletePlannedCashflow {
	return &DeletePlannedCashflow{store: store}
}

func (uc *DeletePlannedCashflow) Execute(ctx context.Context, userID ids.UserID, id ids.PlannedCashflowID) error {
	if userID.IsZero() || id.IsZero() {
		return fmt.Errorf("user id and plan id are required")
	}
	return uc.store.DeletePlanned(ctx, userID, id)
}

func plannedToDTO(p domain.PlannedCashflow) PlanItemDTO {
	return PlanItemDTO{
		ID:          p.ID.String(),
		Kind:        string(p.Kind),
		Title:       p.Title,
		AmountCents: p.AmountCents,
		Currency:    "RUB",
		Interval:    string(p.Interval),
		NextDate:    p.NextDate.Format("2006-01-02"),
		Source:      "plan",
	}
}
