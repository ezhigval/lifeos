package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

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

func (r *Repository) SaveDebt(ctx context.Context, debt domain.Debt) error {
	return r.queries(ctx).InsertDebt(ctx, db.InsertDebtParams{
		ID:          pgconv.DebtID(debt.ID),
		UserID:      pgconv.UserID(debt.UserID),
		Creditor:    debt.Creditor,
		AmountCents: debt.AmountCents,
		PaidCents:   debt.PaidCents,
		DueDate:     pgconv.DatePtr(debt.DueDate),
		Status:      string(debt.Status),
		CreatedAt:   pgconv.TimestamptzValue(debt.CreatedAt),
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
	return r.queries(ctx).UpdateDebt(ctx, db.UpdateDebtParams{
		ID:        pgconv.DebtID(debt.ID),
		UserID:    pgconv.UserID(debt.UserID),
		PaidCents: debt.PaidCents,
		Status:    string(debt.Status),
	})
}

func mapDebt(row db.Debt) domain.Debt {
	return domain.Debt{
		ID:          pgconv.FromDebtID(row.ID),
		UserID:      pgconv.FromUserID(row.UserID),
		Creditor:    row.Creditor,
		AmountCents: row.AmountCents,
		PaidCents:   row.PaidCents,
		DueDate:     pgconv.FromDate(row.DueDate),
		Status:      domain.DebtStatus(row.Status),
		CreatedAt:   row.CreatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
