package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/knowledge/domain"
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

func (r *Repository) Save(ctx context.Context, note domain.Note) error {
	tags := note.Tags
	if tags == nil {
		tags = []string{}
	}
	return r.queries(ctx).InsertNote(ctx, db.InsertNoteParams{
		ID:        pgconv.NoteID(note.ID),
		UserID:    pgconv.UserID(note.UserID),
		Body:      note.Body,
		Tags:      tags,
		CreatedAt: pgconv.TimestamptzValue(note.CreatedAt),
		UpdatedAt: pgconv.TimestamptzValue(note.UpdatedAt),
	})
}

func (r *Repository) ListRecent(ctx context.Context, userID ids.UserID, limit int32) ([]domain.Note, error) {
	rows, err := r.queries(ctx).ListRecentNotesByUser(ctx, db.ListRecentNotesByUserParams{
		UserID: pgconv.UserID(userID),
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	out := make([]domain.Note, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRecentRow(row))
	}
	return out, nil
}

func (r *Repository) ListByTag(ctx context.Context, userID ids.UserID, tag string, limit int32) ([]domain.Note, error) {
	rows, err := r.queries(ctx).ListNotesByTag(ctx, db.ListNotesByTagParams{
		UserID:      pgconv.UserID(userID),
		Tag:         tag,
		ResultLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list notes by tag: %w", err)
	}
	out := make([]domain.Note, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTagRow(row))
	}
	return out, nil
}

func (r *Repository) Search(ctx context.Context, userID ids.UserID, query string, limit int32) ([]domain.Note, error) {
	rows, err := r.queries(ctx).SearchNotesByUser(ctx, db.SearchNotesByUserParams{
		UserID:      pgconv.UserID(userID),
		Query:       pgtype.Text{String: query, Valid: true},
		ResultLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}
	out := make([]domain.Note, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSearchRow(row))
	}
	return out, nil
}

func (r *Repository) Delete(ctx context.Context, userID ids.UserID, noteID ids.NoteID) (domain.Note, error) {
	row, err := r.queries(ctx).DeleteNoteByUser(ctx, db.DeleteNoteByUserParams{
		NoteID: pgconv.NoteID(noteID),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("delete note: %w", err)
	}
	return mapDeleteRow(row), nil
}

func mapRecentRow(row db.ListRecentNotesByUserRow) domain.Note {
	return mapFields(row.ID, row.UserID, row.Body, row.Tags, row.CreatedAt, row.UpdatedAt)
}

func mapTagRow(row db.ListNotesByTagRow) domain.Note {
	return mapFields(row.ID, row.UserID, row.Body, row.Tags, row.CreatedAt, row.UpdatedAt)
}

func mapSearchRow(row db.SearchNotesByUserRow) domain.Note {
	return mapFields(row.ID, row.UserID, row.Body, row.Tags, row.CreatedAt, row.UpdatedAt)
}

func mapDeleteRow(row db.DeleteNoteByUserRow) domain.Note {
	return mapFields(row.ID, row.UserID, row.Body, row.Tags, row.CreatedAt, row.UpdatedAt)
}

func mapFields(id, userID pgtype.UUID, body string, tags []string, createdAt, updatedAt pgtype.Timestamptz) domain.Note {
	if tags == nil {
		tags = []string{}
	}
	return domain.Note{
		ID:        pgconv.FromNoteID(id),
		UserID:    pgconv.FromUserID(userID),
		Body:      body,
		Tags:      tags,
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
