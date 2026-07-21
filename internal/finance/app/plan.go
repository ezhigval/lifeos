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
	GetPlanned(ctx context.Context, userID ids.UserID, id ids.PlannedCashflowID) (domain.PlannedCashflow, error)
	UpdatePlannedNextDate(ctx context.Context, item domain.PlannedCashflow) error
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

// CompletePlanOccurrence marks one occurrence: deletes once-items, advances recurring,
// and posts a matching income/expense transaction when recorders are wired.
type CompletePlanOccurrence struct {
	store      PlannedCashflowStore
	events     EventLog
	transactor Transactor
	income     *RecordIncome
	expense    *RecordExpense
	now        func() time.Time
}

func NewCompletePlanOccurrence(
	store PlannedCashflowStore,
	events EventLog,
	transactor Transactor,
	income *RecordIncome,
	expense *RecordExpense,
) *CompletePlanOccurrence {
	return &CompletePlanOccurrence{
		store: store, events: events, transactor: transactor,
		income: income, expense: expense,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CompletePlanResult struct {
	Deleted      bool
	Item         *PlanItemDTO
	Posted       bool
	PostedCents  int64
	PostedKind   string // income|expense
}

func (uc *CompletePlanOccurrence) Execute(ctx context.Context, userID ids.UserID, id ids.PlannedCashflowID, source events.Source) (CompletePlanResult, error) {
	if userID.IsZero() || id.IsZero() {
		return CompletePlanResult{}, fmt.Errorf("user id and plan id are required")
	}
	item, err := uc.store.GetPlanned(ctx, userID, id)
	if err != nil {
		return CompletePlanResult{}, err
	}
	now := uc.now()
	shouldDelete := item.AdvanceOccurrence(now)
	var out CompletePlanResult
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if shouldDelete {
			if err := uc.store.DeletePlanned(txCtx, userID, id); err != nil {
				return err
			}
			out.Deleted = true
			return uc.events.Append(txCtx, events.Record{
				UserID:        userID,
				AggregateType: "planned_cashflow",
				AggregateID:   id.UUID(),
				EventType:     "PlannedCashflowCompleted",
				Payload:       map[string]any{"deleted": true, "title": item.Title},
				Source:        source,
				OccurredAt:    now,
			})
		}
		if err := uc.store.UpdatePlannedNextDate(txCtx, item); err != nil {
			return err
		}
		dto := plannedToDTO(item)
		out.Item = &dto
		return uc.events.Append(txCtx, events.Record{
			UserID:        userID,
			AggregateType: "planned_cashflow",
			AggregateID:   id.UUID(),
			EventType:     "PlannedCashflowAdvanced",
			Payload: map[string]any{
				"title":     item.Title,
				"next_date": item.NextDate.Format("2006-01-02"),
				"interval":  item.Interval,
			},
			Source:     source,
			OccurredAt: now,
		})
	})
	if err != nil {
		return CompletePlanResult{}, fmt.Errorf("complete plan occurrence: %w", err)
	}

	if posted, err := postPlannedCashflow(ctx, uc.income, uc.expense, item, source); err != nil {
		return out, fmt.Errorf("post planned cashflow: %w", err)
	} else if posted {
		out.Posted = true
		out.PostedCents = item.AmountCents
		out.PostedKind = string(item.Kind)
	}
	return out, nil
}

func postPlannedCashflow(
	ctx context.Context,
	income *RecordIncome,
	expense *RecordExpense,
	item domain.PlannedCashflow,
	source events.Source,
) (bool, error) {
	desc := "План · " + item.Title
	switch item.Kind {
	case domain.PlanKindIncome:
		if income == nil {
			return false, nil
		}
		_, err := income.Execute(ctx, RecordIncomeInput{
			UserID:      item.UserID,
			AmountCents: item.AmountCents,
			Currency:    "RUB",
			Description: desc,
			Source:      source,
		})
		return err == nil, err
	case domain.PlanKindExpense:
		if expense == nil {
			return false, nil
		}
		_, err := expense.Execute(ctx, RecordExpenseInput{
			UserID:       item.UserID,
			AmountCents:  item.AmountCents,
			Currency:     "RUB",
			CategoryName: "План",
			Description:  desc,
			Source:       source,
		})
		return err == nil, err
	default:
		return false, nil
	}
}

// AdvanceOverduePlans rolls/deletes plan rows with next_date strictly before today
// and posts one ledger entry per rolled occurrence.
type AdvanceOverduePlans struct {
	store      PlannedCashflowStore
	events     EventLog
	transactor Transactor
	income     *RecordIncome
	expense    *RecordExpense
	now        func() time.Time
}

func NewAdvanceOverduePlans(
	store PlannedCashflowStore,
	events EventLog,
	transactor Transactor,
	income *RecordIncome,
	expense *RecordExpense,
) *AdvanceOverduePlans {
	return &AdvanceOverduePlans{
		store: store, events: events, transactor: transactor,
		income: income, expense: expense,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type AdvanceOverdueResult struct {
	Advanced int
	Deleted  int
	Posted   int
}

func (uc *AdvanceOverduePlans) Execute(ctx context.Context, userID ids.UserID, source events.Source) (AdvanceOverdueResult, error) {
	if userID.IsZero() {
		return AdvanceOverdueResult{}, fmt.Errorf("user id is required")
	}
	items, err := uc.store.ListPlanned(ctx, userID)
	if err != nil {
		return AdvanceOverdueResult{}, err
	}
	now := uc.now()
	var result AdvanceOverdueResult
	for _, item := range items {
		snapshot := item // amount/kind/title before date roll
		changed, shouldDelete := item.AdvanceIfOverdue(now)
		if !changed {
			continue
		}
		err := uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			if shouldDelete {
				if err := uc.store.DeletePlanned(txCtx, userID, item.ID); err != nil {
					return err
				}
				result.Deleted++
				return uc.events.Append(txCtx, events.Record{
					UserID:        userID,
					AggregateType: "planned_cashflow",
					AggregateID:   item.ID.UUID(),
					EventType:     "PlannedCashflowAutoDeleted",
					Payload:       map[string]any{"title": item.Title},
					Source:        source,
					OccurredAt:    now,
				})
			}
			if err := uc.store.UpdatePlannedNextDate(txCtx, item); err != nil {
				return err
			}
			result.Advanced++
			return uc.events.Append(txCtx, events.Record{
				UserID:        userID,
				AggregateType: "planned_cashflow",
				AggregateID:   item.ID.UUID(),
				EventType:     "PlannedCashflowAutoAdvanced",
				Payload: map[string]any{
					"title":     item.Title,
					"next_date": item.NextDate.Format("2006-01-02"),
				},
				Source:     source,
				OccurredAt: now,
			})
		})
		if err != nil {
			return result, err
		}
		if posted, postErr := postPlannedCashflow(ctx, uc.income, uc.expense, snapshot, source); postErr != nil {
			return result, fmt.Errorf("post overdue planned cashflow: %w", postErr)
		} else if posted {
			result.Posted++
		}
	}
	return result, nil
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
