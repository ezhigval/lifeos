package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/finance/app"
	"github.com/valentinezhov/lifeos/internal/finance/domain"
	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByName(ctx context.Context, userID ids.UserID, name string, kind domain.Kind) (domain.Category, error) {
	row, err := r.queries(ctx).GetFinanceCategoryByName(ctx, db.GetFinanceCategoryByNameParams{
		UserID: pgconv.UserID(userID),
		Name:   name,
		Kind:   string(kind),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Category{}, domain.ErrCategoryNotFound
	}
	if err != nil {
		return domain.Category{}, fmt.Errorf("get category: %w", err)
	}
	return domain.Category{
		ID:        pgconv.FromCategoryID(row.ID),
		UserID:    pgconv.FromUserID(row.UserID),
		Name:      row.Name,
		Kind:      domain.Kind(row.Kind),
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *Repository) SaveCategory(ctx context.Context, cat domain.Category) error {
	return r.queries(ctx).InsertFinanceCategory(ctx, db.InsertFinanceCategoryParams{
		ID:        pgconv.CategoryID(cat.ID),
		UserID:    pgconv.UserID(cat.UserID),
		Name:      cat.Name,
		Kind:      string(cat.Kind),
		CreatedAt: pgconv.TimestamptzValue(cat.CreatedAt),
	})
}

func (r *Repository) SaveTransaction(ctx context.Context, tx domain.Transaction) error {
	return r.queries(ctx).InsertFinanceTransaction(ctx, db.InsertFinanceTransactionParams{
		ID:          pgconv.TransactionID(tx.ID),
		UserID:      pgconv.UserID(tx.UserID),
		CategoryID:  pgconv.CategoryID(tx.CategoryID),
		Kind:        string(tx.Kind),
		AmountCents: tx.Money.AmountCents,
		Currency:    tx.Money.Currency,
		Description: pgconv.Text(&tx.Description),
		OccurredAt:  pgconv.TimestamptzValue(tx.OccurredAt),
		CreatedAt:   pgconv.TimestamptzValue(tx.CreatedAt),
	})
}

func (r *Repository) SumIncomeBetween(ctx context.Context, userID ids.UserID, from, to time.Time) (int64, error) {
	total, err := r.queries(ctx).SumIncomeBetween(ctx, db.SumIncomeBetweenParams{
		UserID:       pgconv.UserID(userID),
		OccurredAt:   pgconv.TimestamptzValue(from),
		OccurredAt_2: pgconv.TimestamptzValue(to),
	})
	if err != nil {
		return 0, fmt.Errorf("sum income: %w", err)
	}
	return total, nil
}

func (r *Repository) SumExpenseBetween(ctx context.Context, userID ids.UserID, from, to time.Time) (int64, error) {
	total, err := r.queries(ctx).SumExpenseBetween(ctx, db.SumExpenseBetweenParams{
		UserID:       pgconv.UserID(userID),
		OccurredAt:   pgconv.TimestamptzValue(from),
		OccurredAt_2: pgconv.TimestamptzValue(to),
	})
	if err != nil {
		return 0, fmt.Errorf("sum expense: %w", err)
	}
	return total, nil
}

func (r *Repository) SumExpensesByCategoryBetween(ctx context.Context, userID ids.UserID, from, to time.Time) ([]app.CategoryExpenseTotal, error) {
	rows, err := r.queries(ctx).SumExpensesByCategoryBetween(ctx, db.SumExpensesByCategoryBetweenParams{
		UserID:       pgconv.UserID(userID),
		OccurredAt:   pgconv.TimestamptzValue(from),
		OccurredAt_2: pgconv.TimestamptzValue(to),
	})
	if err != nil {
		return nil, fmt.Errorf("sum expenses by category: %w", err)
	}
	out := make([]app.CategoryExpenseTotal, 0, len(rows))
	for _, row := range rows {
		out = append(out, app.CategoryExpenseTotal{
			Name:        row.Name,
			AmountCents: row.AmountCents,
		})
	}
	return out, nil
}

func (r *Repository) SaveDebt(ctx context.Context, debt domain.Debt) error {
	interval := debt.InstallmentInterval
	if interval == "" {
		interval = "none"
	}
	return r.queries(ctx).InsertDebt(ctx, db.InsertDebtParams{
		ID:                  pgconv.DebtID(debt.ID),
		UserID:              pgconv.UserID(debt.UserID),
		Creditor:            debt.Creditor,
		AmountCents:         debt.AmountCents,
		PaidCents:           debt.PaidCents,
		DueDate:             pgconv.DatePtr(debt.DueDate),
		Status:              string(debt.Status),
		CreatedAt:           pgconv.TimestamptzValue(debt.CreatedAt),
		InstallmentCents:    debt.InstallmentCents,
		InstallmentInterval: interval,
		NextPaymentDate:     pgconv.DatePtr(debt.NextPaymentDate),
	})
}

func (r *Repository) ListOpen(ctx context.Context, userID ids.UserID) ([]domain.Debt, error) {
	rows, err := r.queries(ctx).ListOpenDebts(ctx, pgconv.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("list debts: %w", err)
	}
	out := make([]domain.Debt, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapDebt(row))
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, userID ids.UserID, debtID ids.DebtID) (domain.Debt, error) {
	row, err := r.queries(ctx).GetDebtByID(ctx, db.GetDebtByIDParams{
		ID:     pgconv.DebtID(debtID),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Debt{}, domain.ErrDebtNotFound
	}
	if err != nil {
		return domain.Debt{}, fmt.Errorf("get debt: %w", err)
	}
	return mapDebt(row), nil
}

func (r *Repository) FindOpenByCreditor(ctx context.Context, userID ids.UserID, creditor string) (domain.Debt, error) {
	row, err := r.queries(ctx).FindOpenDebtByCreditor(ctx, db.FindOpenDebtByCreditorParams{
		UserID:  pgconv.UserID(userID),
		Column2: pgconv.Text(&creditor),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Debt{}, domain.ErrDebtNotFound
	}
	if err != nil {
		return domain.Debt{}, fmt.Errorf("find debt: %w", err)
	}
	return mapDebt(row), nil
}

func (r *Repository) UpdateDebt(ctx context.Context, debt domain.Debt) error {
	interval := debt.InstallmentInterval
	if interval == "" {
		interval = "none"
	}
	return r.queries(ctx).UpdateDebt(ctx, db.UpdateDebtParams{
		ID:                  pgconv.DebtID(debt.ID),
		UserID:              pgconv.UserID(debt.UserID),
		PaidCents:           debt.PaidCents,
		Status:              string(debt.Status),
		InstallmentCents:    debt.InstallmentCents,
		InstallmentInterval: interval,
		NextPaymentDate:     pgconv.DatePtr(debt.NextPaymentDate),
	})
}

func (r *Repository) SavePlanned(ctx context.Context, item domain.PlannedCashflow) error {
	return r.queries(ctx).InsertPlannedCashflow(ctx, db.InsertPlannedCashflowParams{
		ID:          pgconv.PlannedCashflowID(item.ID),
		UserID:      pgconv.UserID(item.UserID),
		Kind:        string(item.Kind),
		Title:       item.Title,
		AmountCents: item.AmountCents,
		Interval:    string(item.Interval),
		NextDate:    pgconv.Date(item.NextDate),
		DebtID:      pgconv.DebtIDPtr(item.DebtID),
		CreatedAt:   pgconv.TimestamptzValue(item.CreatedAt),
	})
}

func (r *Repository) GetPlanned(ctx context.Context, userID ids.UserID, id ids.PlannedCashflowID) (domain.PlannedCashflow, error) {
	row, err := r.queries(ctx).GetPlannedCashflowByUser(ctx, db.GetPlannedCashflowByUserParams{
		ID:     pgconv.PlannedCashflowID(id),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PlannedCashflow{}, domain.ErrPlanNotFound
	}
	if err != nil {
		return domain.PlannedCashflow{}, fmt.Errorf("get planned cashflow: %w", err)
	}
	return mapPlanned(row), nil
}

func (r *Repository) UpdatePlannedNextDate(ctx context.Context, item domain.PlannedCashflow) error {
	return r.queries(ctx).UpdatePlannedCashflowNextDate(ctx, db.UpdatePlannedCashflowNextDateParams{
		ID:        pgconv.PlannedCashflowID(item.ID),
		UserID:    pgconv.UserID(item.UserID),
		NextDate:  pgconv.Date(item.NextDate),
		UpdatedAt: pgconv.TimestamptzValue(item.UpdatedAt),
	})
}

func (r *Repository) ListPlanned(ctx context.Context, userID ids.UserID) ([]domain.PlannedCashflow, error) {
	rows, err := r.queries(ctx).ListPlannedCashflowsByUser(ctx, pgconv.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("list planned cashflows: %w", err)
	}
	out := make([]domain.PlannedCashflow, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapPlanned(row))
	}
	return out, nil
}

func (r *Repository) DeletePlanned(ctx context.Context, userID ids.UserID, id ids.PlannedCashflowID) error {
	_, err := r.queries(ctx).DeletePlannedCashflowByUser(ctx, db.DeletePlannedCashflowByUserParams{
		ID:     pgconv.PlannedCashflowID(id),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrPlanNotFound
	}
	if err != nil {
		return fmt.Errorf("delete planned cashflow: %w", err)
	}
	return nil
}

func mapPlanned(row db.PlannedCashflow) domain.PlannedCashflow {
	next := time.Time{}
	if row.NextDate.Valid {
		next = row.NextDate.Time
	}
	return domain.PlannedCashflow{
		ID:          pgconv.FromPlannedCashflowID(row.ID),
		UserID:      pgconv.FromUserID(row.UserID),
		Kind:        domain.PlanKind(row.Kind),
		Title:       row.Title,
		AmountCents: row.AmountCents,
		Interval:    domain.PlanInterval(row.Interval),
		NextDate:    next,
		DebtID:      pgconv.FromDebtIDPtr(row.DebtID),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func mapDebt(row db.Debt) domain.Debt {
	interval := row.InstallmentInterval
	if interval == "" {
		interval = "none"
	}
	return domain.Debt{
		ID:                  pgconv.FromDebtID(row.ID),
		UserID:              pgconv.FromUserID(row.UserID),
		Creditor:            row.Creditor,
		AmountCents:         row.AmountCents,
		PaidCents:           row.PaidCents,
		DueDate:             pgconv.FromDate(row.DueDate),
		InstallmentCents:    row.InstallmentCents,
		InstallmentInterval: interval,
		NextPaymentDate:     pgconv.FromDate(row.NextPaymentDate),
		Status:              domain.DebtStatus(row.Status),
		CreatedAt:           row.CreatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
