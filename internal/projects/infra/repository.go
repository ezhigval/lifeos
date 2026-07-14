package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
	"github.com/valentinezhov/lifeos/internal/projects/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, project domain.Project) error {
	return r.queries(ctx).InsertProject(ctx, db.InsertProjectParams{
		ID:           pgconv.ProjectID(project.ID),
		UserID:       pgconv.UserID(project.UserID),
		Name:         project.Name,
		Description:  pgconv.Text(project.Description),
		Status:       string(project.Status),
		Outcome:      pgconv.Text(project.Outcome),
		TargetValue:  decimalToNumeric(project.TargetValue),
		CurrentValue: decimalToNumericValue(project.CurrentValue),
		Unit:         pgconv.Text(project.Unit),
		TargetDate:   pgconv.DatePtr(project.TargetDate),
		CreatedAt:    pgconv.TimestamptzValue(project.CreatedAt),
		UpdatedAt:    pgconv.TimestamptzValue(project.UpdatedAt),
	})
}

func (r *Repository) SetSpheres(ctx context.Context, projectID ids.ProjectID, sphereIDs []ids.SphereID) error {
	q := r.queries(ctx)
	if err := q.DeleteProjectSpheres(ctx, pgconv.ProjectID(projectID)); err != nil {
		return err
	}
	for _, sid := range sphereIDs {
		if err := q.InsertProjectSphere(ctx, db.InsertProjectSphereParams{
			ProjectID: pgconv.ProjectID(projectID),
			SphereID:  pgconv.SphereID(sid),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) LoadSphereIDs(ctx context.Context, projectID ids.ProjectID) ([]ids.SphereID, error) {
	rows, err := r.queries(ctx).ListSphereIDsByProject(ctx, pgconv.ProjectID(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]ids.SphereID, 0, len(rows))
	for _, row := range rows {
		out = append(out, pgconv.FromSphereID(row))
	}
	return out, nil
}

func (r *Repository) FindByName(ctx context.Context, userID ids.UserID, name string) (domain.Project, error) {
	row, err := r.queries(ctx).FindProjectByName(ctx, db.FindProjectByNameParams{
		UserID: pgconv.UserID(userID),
		Lower:  name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("find project: %w", err)
	}
	return r.attachSpheres(ctx, mapProjectRow(row.ID, row.UserID, row.Name, row.Description, row.Status, row.Outcome, row.TargetValue, row.CurrentValue, row.Unit, row.TargetDate, row.CreatedAt, row.UpdatedAt))
}

func (r *Repository) ListActive(ctx context.Context, userID ids.UserID) ([]domain.Project, error) {
	rows, err := r.queries(ctx).ListActiveProjects(ctx, pgconv.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	out := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		p, err := r.attachSpheres(ctx, mapProjectRow(row.ID, row.UserID, row.Name, row.Description, row.Status, row.Outcome, row.TargetValue, row.CurrentValue, row.Unit, row.TargetDate, row.CreatedAt, row.UpdatedAt))
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *Repository) ListBySphere(ctx context.Context, userID ids.UserID, sphereID ids.SphereID) ([]domain.Project, error) {
	rows, err := r.queries(ctx).ListProjectsBySphere(ctx, db.ListProjectsBySphereParams{
		UserID:   pgconv.UserID(userID),
		SphereID: pgconv.SphereID(sphereID),
	})
	if err != nil {
		return nil, fmt.Errorf("list projects by sphere: %w", err)
	}
	out := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		p, err := r.attachSpheres(ctx, mapProjectRow(row.ID, row.UserID, row.Name, row.Description, row.Status, row.Outcome, row.TargetValue, row.CurrentValue, row.Unit, row.TargetDate, row.CreatedAt, row.UpdatedAt))
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *Repository) Exists(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (bool, error) {
	ok, err := r.queries(ctx).ProjectExists(ctx, db.ProjectExistsParams{
		ID:     pgconv.ProjectID(projectID),
		UserID: pgconv.UserID(userID),
	})
	if err != nil {
		return false, fmt.Errorf("project exists: %w", err)
	}
	return ok, nil
}

func (r *Repository) AllExist(ctx context.Context, userID ids.UserID, projectIDs []ids.ProjectID) (bool, error) {
	if len(projectIDs) == 0 {
		return true, nil
	}
	uuids := make([]pgtype.UUID, 0, len(projectIDs))
	for _, id := range projectIDs {
		uuids = append(uuids, pgconv.ProjectID(id))
	}
	ok, err := r.queries(ctx).ProjectsExist(ctx, db.ProjectsExistParams{
		UserID:     pgconv.UserID(userID),
		ProjectIds: uuids,
		Expected:   int32(len(projectIDs)),
	})
	if err != nil {
		return false, fmt.Errorf("projects exist: %w", err)
	}
	return ok, nil
}

func (r *Repository) GetByID(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (domain.Project, error) {
	row, err := r.queries(ctx).GetProjectByID(ctx, db.GetProjectByIDParams{
		ID:     pgconv.ProjectID(projectID),
		UserID: pgconv.UserID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("get project: %w", err)
	}
	return r.attachSpheres(ctx, mapProjectRow(row.ID, row.UserID, row.Name, row.Description, row.Status, row.Outcome, row.TargetValue, row.CurrentValue, row.Unit, row.TargetDate, row.CreatedAt, row.UpdatedAt))
}

func (r *Repository) UpdateStatus(ctx context.Context, project domain.Project) error {
	return r.queries(ctx).UpdateProjectStatus(ctx, db.UpdateProjectStatusParams{
		ID:     pgconv.ProjectID(project.ID),
		UserID: pgconv.UserID(project.UserID),
		Status: string(project.Status),
	})
}

func (r *Repository) attachSpheres(ctx context.Context, p domain.Project) (domain.Project, error) {
	ids, err := r.LoadSphereIDs(ctx, p.ID)
	if err != nil {
		return domain.Project{}, err
	}
	p.SphereIDs = ids
	return p, nil
}

func mapProjectRow(
	id, userID pgtype.UUID,
	name string,
	description pgtype.Text,
	status string,
	outcome pgtype.Text,
	targetValue, currentValue pgtype.Numeric,
	unit pgtype.Text,
	targetDate pgtype.Date,
	createdAt, updatedAt pgtype.Timestamptz,
) domain.Project {
	return domain.Project{
		ID:           pgconv.FromProjectID(id),
		UserID:       pgconv.FromUserID(userID),
		Name:         name,
		Description:  pgconv.FromText(description),
		Outcome:      pgconv.FromText(outcome),
		Status:       domain.Status(status),
		TargetValue:  numericToDecimalPtr(targetValue),
		CurrentValue: numericToDecimalValue(currentValue),
		Unit:         pgconv.FromText(unit),
		TargetDate:   pgconv.FromDate(targetDate),
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}
}

func (r *Repository) queries(ctx context.Context) *db.Queries {
	if tx, ok := platformpostgres.TxFromContext(ctx); ok {
		return db.New(tx)
	}
	return db.New(r.pool)
}

type ProjectReader struct {
	*Repository
}

func NewProjectReader(pool *pgxpool.Pool) *ProjectReader {
	return &ProjectReader{Repository: NewRepository(pool)}
}

func (r *ProjectReader) Exists(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (bool, error) {
	return r.Repository.Exists(ctx, userID, projectID)
}

func (r *ProjectReader) AllExist(ctx context.Context, userID ids.UserID, projectIDs []ids.ProjectID) (bool, error) {
	return r.Repository.AllExist(ctx, userID, projectIDs)
}

// unused import guard for uuid in array conversion if needed
var _ = uuid.Nil

func decimalToNumeric(d *decimal.Decimal) pgtype.Numeric {
	if d == nil {
		return pgtype.Numeric{}
	}
	return decimalToNumericValue(*d)
}

func decimalToNumericValue(d decimal.Decimal) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(d.String())
	return n
}

func numericToDecimalPtr(n pgtype.Numeric) *decimal.Decimal {
	if !n.Valid {
		return nil
	}
	d := numericToDecimalValue(n)
	return &d
}

func numericToDecimalValue(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	f, _ := n.Float64Value()
	return decimal.NewFromFloat(f.Float64)
}
