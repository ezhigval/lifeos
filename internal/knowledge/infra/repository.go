package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	targetType, targetID := noteTargetParams(note)
	return r.queries(ctx).InsertNote(ctx, db.InsertNoteParams{
		ID:         pgconv.NoteID(note.ID),
		UserID:     pgconv.UserID(note.UserID),
		Body:       note.Body,
		Tags:       tags,
		TargetType: targetType,
		TargetID:   targetID,
		CreatedAt:  pgconv.TimestamptzValue(note.CreatedAt),
		UpdatedAt:  pgconv.TimestamptzValue(note.UpdatedAt),
	})
}

func noteTargetParams(note domain.Note) (pgtype.Text, pgtype.UUID) {
	var tt pgtype.Text
	var tid pgtype.UUID
	if note.TargetType != nil {
		tt = pgtype.Text{String: string(*note.TargetType), Valid: true}
	}
	if note.TargetID != nil {
		tid = pgconv.UUID(*note.TargetID)
	}
	return tt, tid
}

func targetFromRow(targetType pgtype.Text, targetID pgtype.UUID) (*domain.TargetType, *uuid.UUID) {
	var tt *domain.TargetType
	if targetType.Valid && targetType.String != "" {
		v := domain.TargetType(targetType.String)
		tt = &v
	}
	var tid *uuid.UUID
	if targetID.Valid {
		v := pgconv.FromUUID(targetID)
		tid = &v
	}
	return tt, tid
}

func (r *Repository) GetByID(ctx context.Context, userID ids.UserID, noteID ids.NoteID) (domain.Note, error) {
	row, err := r.queries(ctx).GetNoteByID(ctx, db.GetNoteByIDParams{
		ID:     pgconv.NoteID(noteID),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("get note: %w", err)
	}
	tt, tid := targetFromRow(row.TargetType, row.TargetID)
	return domain.Note{
		ID:         pgconv.FromNoteID(row.ID),
		UserID:     pgconv.FromUserID(row.UserID),
		Body:       row.Body,
		Tags:       row.Tags,
		TargetType: tt,
		TargetID:   tid,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}, nil
}

func (r *Repository) ListByTarget(ctx context.Context, userID ids.UserID, targetType domain.TargetType, targetID uuid.UUID) ([]domain.Note, error) {
	rows, err := r.queries(ctx).ListNotesByTarget(ctx, db.ListNotesByTargetParams{
		UserID:     pgconv.UserID(userID),
		TargetType: string(targetType),
		TargetID:   pgconv.UUID(targetID),
	})
	if err != nil {
		return nil, fmt.Errorf("list notes by target: %w", err)
	}
	out := make([]domain.Note, 0, len(rows))
	for _, row := range rows {
		tt, tid := targetFromRow(row.TargetType, row.TargetID)
		out = append(out, domain.Note{
			ID:         pgconv.FromNoteID(row.ID),
			UserID:     pgconv.FromUserID(row.UserID),
			Body:       row.Body,
			Tags:       row.Tags,
			TargetType: tt,
			TargetID:   tid,
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		})
	}
	return out, nil
}

func (r *Repository) UpdateBody(ctx context.Context, userID ids.UserID, noteID ids.NoteID, body string, now time.Time) (domain.Note, error) {
	row, err := r.queries(ctx).UpdateNoteBody(ctx, db.UpdateNoteBodyParams{
		ID:        pgconv.NoteID(noteID),
		UserID:    pgconv.UserID(userID),
		Body:      body,
		UpdatedAt: pgconv.TimestamptzValue(now),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Note{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("update note: %w", err)
	}
	tt, tid := targetFromRow(row.TargetType, row.TargetID)
	return domain.Note{
		ID:         pgconv.FromNoteID(row.ID),
		UserID:     pgconv.FromUserID(row.UserID),
		Body:       row.Body,
		Tags:       row.Tags,
		TargetType: tt,
		TargetID:   tid,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}, nil
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

func (r *Repository) ListCreatedBetween(ctx context.Context, userID ids.UserID, from, to time.Time) ([]domain.Note, error) {
	rows, err := r.queries(ctx).ListNotesCreatedBetween(ctx, db.ListNotesCreatedBetweenParams{
		UserID:     pgconv.UserID(userID),
		CreatedAt:  pgconv.TimestamptzValue(from),
		CreatedAt_: pgconv.TimestamptzValue(to),
	})
	if err != nil {
		return nil, fmt.Errorf("list notes created between: %w", err)
	}
	out := make([]domain.Note, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapBetweenRow(row))
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
	tt, tid := targetFromRow(row.TargetType, row.TargetID)
	return mapFieldsWithTarget(row.ID, row.UserID, row.Body, row.Tags, tt, tid, row.CreatedAt, row.UpdatedAt)
}

func mapTagRow(row db.ListNotesByTagRow) domain.Note {
	tt, tid := targetFromRow(row.TargetType, row.TargetID)
	return mapFieldsWithTarget(row.ID, row.UserID, row.Body, row.Tags, tt, tid, row.CreatedAt, row.UpdatedAt)
}

func mapSearchRow(row db.SearchNotesByUserRow) domain.Note {
	tt, tid := targetFromRow(row.TargetType, row.TargetID)
	return mapFieldsWithTarget(row.ID, row.UserID, row.Body, row.Tags, tt, tid, row.CreatedAt, row.UpdatedAt)
}

func mapDeleteRow(row db.DeleteNoteByUserRow) domain.Note {
	tt, tid := targetFromRow(row.TargetType, row.TargetID)
	return mapFieldsWithTarget(row.ID, row.UserID, row.Body, row.Tags, tt, tid, row.CreatedAt, row.UpdatedAt)
}

func mapBetweenRow(row db.ListRecentNotesByUserRow) domain.Note {
	tt, tid := targetFromRow(row.TargetType, row.TargetID)
	return mapFieldsWithTarget(row.ID, row.UserID, row.Body, row.Tags, tt, tid, row.CreatedAt, row.UpdatedAt)
}

func mapFieldsWithTarget(id, userID pgtype.UUID, body string, tags []string, tt *domain.TargetType, tid *uuid.UUID, createdAt, updatedAt pgtype.Timestamptz) domain.Note {
	if tags == nil {
		tags = []string{}
	}
	return domain.Note{
		ID:         pgconv.FromNoteID(id),
		UserID:     pgconv.FromUserID(userID),
		Body:       body,
		Tags:       tags,
		TargetType: tt,
		TargetID:   tid,
		CreatedAt:  createdAt.Time,
		UpdatedAt:  updatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}
