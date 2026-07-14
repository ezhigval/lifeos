package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/valentinezhov/lifeos/internal/career/domain"
	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
)

func (r *Repository) SaveSkill(ctx context.Context, skill domain.Skill) error {
	return r.queries(ctx).InsertSkill(ctx, db.InsertSkillParams{
		ID:        pgconv.SkillID(skill.ID),
		UserID:    pgconv.UserID(skill.UserID),
		Name:      skill.Name,
		Level:     skill.Level,
		CreatedAt: pgconv.TimestamptzValue(skill.CreatedAt),
		UpdatedAt: pgconv.TimestamptzValue(skill.UpdatedAt),
	})
}

func (r *Repository) ListRecentSkills(ctx context.Context, userID ids.UserID, limit int32) ([]domain.Skill, error) {
	rows, err := r.queries(ctx).ListRecentSkillsByUser(ctx, db.ListRecentSkillsByUserParams{
		UserID: pgconv.UserID(userID),
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	out := make([]domain.Skill, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSkillRow(row))
	}
	return out, nil
}

func (r *Repository) SearchSkills(ctx context.Context, userID ids.UserID, query string, limit int32) ([]domain.Skill, error) {
	rows, err := r.queries(ctx).SearchSkillsByUser(ctx, db.SearchSkillsByUserParams{
		UserID:      pgconv.UserID(userID),
		Query:       pgtype.Text{String: query, Valid: true},
		ResultLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search skills: %w", err)
	}
	out := make([]domain.Skill, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSkillRow(row))
	}
	return out, nil
}

func (r *Repository) DeleteSkill(ctx context.Context, userID ids.UserID, skillID ids.SkillID) (domain.Skill, error) {
	row, err := r.queries(ctx).DeleteSkillByUser(ctx, db.DeleteSkillByUserParams{
		SkillID: pgconv.SkillID(skillID),
		UserID:  pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Skill{}, domain.ErrSkillNotFound
	}
	if err != nil {
		return domain.Skill{}, fmt.Errorf("delete skill: %w", err)
	}
	return mapSkillRow(row), nil
}

func mapSkillRow(row db.CareerSkill) domain.Skill {
	return domain.Skill{
		ID:        pgconv.FromSkillID(row.ID),
		UserID:    pgconv.FromUserID(row.UserID),
		Name:      row.Name,
		Level:     row.Level,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
