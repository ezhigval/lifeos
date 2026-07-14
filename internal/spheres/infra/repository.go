package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
	"github.com/valentinezhov/lifeos/internal/spheres/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, sphere domain.Sphere) error {
	return r.queries(ctx).InsertSphere(ctx, db.InsertSphereParams{
		ID:        pgconv.SphereID(sphere.ID),
		UserID:    pgconv.UserID(sphere.UserID),
		Name:      sphere.Name,
		SortOrder: sphere.SortOrder,
		CreatedAt: pgconv.TimestamptzValue(sphere.CreatedAt),
		UpdatedAt: pgconv.TimestamptzValue(sphere.UpdatedAt),
	})
}

func (r *Repository) List(ctx context.Context, userID ids.UserID) ([]domain.Sphere, error) {
	rows, err := r.queries(ctx).ListSpheresByUser(ctx, pgconv.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("list spheres: %w", err)
	}
	out := make([]domain.Sphere, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}
	return out, nil
}

func (r *Repository) Get(ctx context.Context, userID ids.UserID, sphereID ids.SphereID) (domain.Sphere, error) {
	row, err := r.queries(ctx).GetSphereByUser(ctx, db.GetSphereByUserParams{
		ID:     pgconv.SphereID(sphereID),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Sphere{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Sphere{}, fmt.Errorf("get sphere: %w", err)
	}
	return mapRow(row), nil
}

func (r *Repository) FindByName(ctx context.Context, userID ids.UserID, name string) (domain.Sphere, error) {
	row, err := r.queries(ctx).FindSphereByName(ctx, db.FindSphereByNameParams{
		UserID: pgconv.UserID(userID),
		Name:   name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Sphere{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Sphere{}, fmt.Errorf("find sphere: %w", err)
	}
	return mapRow(row), nil
}

func (r *Repository) Count(ctx context.Context, userID ids.UserID) (int32, error) {
	count, err := r.queries(ctx).CountSpheresByUser(ctx, pgconv.UserID(userID))
	if err != nil {
		return 0, fmt.Errorf("count spheres: %w", err)
	}
	return count, nil
}

func (r *Repository) Update(ctx context.Context, sphere domain.Sphere) error {
	return r.queries(ctx).UpdateSphere(ctx, db.UpdateSphereParams{
		ID:        pgconv.SphereID(sphere.ID),
		UserID:    pgconv.UserID(sphere.UserID),
		Name:      sphere.Name,
		SortOrder: sphere.SortOrder,
		UpdatedAt: pgconv.TimestamptzValue(sphere.UpdatedAt),
	})
}

func (r *Repository) Delete(ctx context.Context, userID ids.UserID, sphereID ids.SphereID) (domain.Sphere, error) {
	row, err := r.queries(ctx).DeleteSphereByUser(ctx, db.DeleteSphereByUserParams{
		ID:     pgconv.SphereID(sphereID),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Sphere{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Sphere{}, fmt.Errorf("delete sphere: %w", err)
	}
	return mapRow(row), nil
}

func (r *Repository) HasLinkedProjects(ctx context.Context, sphereID ids.SphereID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM project_spheres WHERE sphere_id = $1
		)`, pgconv.SphereID(sphereID)).Scan(&exists)
	if err != nil {
		// project_spheres may not exist yet during phase 1 migrations
		if isUndefinedTable(err) {
			return false, nil
		}
		return false, fmt.Errorf("check linked projects: %w", err)
	}
	return exists, nil
}

func isUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}

func mapRow(row db.LifeSphere) domain.Sphere {
	return domain.Sphere{
		ID:        pgconv.FromSphereID(row.ID),
		UserID:    pgconv.FromUserID(row.UserID),
		Name:      row.Name,
		SortOrder: row.SortOrder,
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
