package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/career/domain"
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

func (r *Repository) Save(ctx context.Context, contact domain.Contact) error {
	return r.queries(ctx).InsertContact(ctx, db.InsertContactParams{
		ID:        pgconv.ContactID(contact.ID),
		UserID:    pgconv.UserID(contact.UserID),
		Name:      contact.Name,
		Company:   contact.Company,
		Role:      contact.Role,
		Notes:     contact.Notes,
		CreatedAt: pgconv.TimestamptzValue(contact.CreatedAt),
		UpdatedAt: pgconv.TimestamptzValue(contact.UpdatedAt),
	})
}

func (r *Repository) ListRecent(ctx context.Context, userID ids.UserID, limit int32) ([]domain.Contact, error) {
	rows, err := r.queries(ctx).ListRecentContactsByUser(ctx, db.ListRecentContactsByUserParams{
		UserID: pgconv.UserID(userID),
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	out := make([]domain.Contact, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}
	return out, nil
}

func (r *Repository) Search(ctx context.Context, userID ids.UserID, query string, limit int32) ([]domain.Contact, error) {
	rows, err := r.queries(ctx).SearchContactsByUser(ctx, db.SearchContactsByUserParams{
		UserID:      pgconv.UserID(userID),
		Query:       pgtype.Text{String: query, Valid: true},
		ResultLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search contacts: %w", err)
	}
	out := make([]domain.Contact, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}
	return out, nil
}

func (r *Repository) Delete(ctx context.Context, userID ids.UserID, contactID ids.ContactID) (domain.Contact, error) {
	row, err := r.queries(ctx).DeleteContactByUser(ctx, db.DeleteContactByUserParams{
		ContactID: pgconv.ContactID(contactID),
		UserID:    pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Contact{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Contact{}, fmt.Errorf("delete contact: %w", err)
	}
	return mapRow(row), nil
}

func mapRow(row db.CareerContact) domain.Contact {
	return domain.Contact{
		ID:        pgconv.FromContactID(row.ID),
		UserID:    pgconv.FromUserID(row.UserID),
		Name:      row.Name,
		Company:   row.Company,
		Role:      row.Role,
		Notes:     row.Notes,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
