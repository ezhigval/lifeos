package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/projects/domain"
)

type ProjectStore interface {
	Save(ctx context.Context, project domain.Project) error
	SetSpheres(ctx context.Context, projectID ids.ProjectID, sphereIDs []ids.SphereID) error
	LoadSphereIDs(ctx context.Context, projectID ids.ProjectID) ([]ids.SphereID, error)
	FindByName(ctx context.Context, userID ids.UserID, name string) (domain.Project, error)
	GetByID(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (domain.Project, error)
	ListActive(ctx context.Context, userID ids.UserID) ([]domain.Project, error)
	ListBySphere(ctx context.Context, userID ids.UserID, sphereID ids.SphereID) ([]domain.Project, error)
	Exists(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (bool, error)
	AllExist(ctx context.Context, userID ids.UserID, projectIDs []ids.ProjectID) (bool, error)
}

type EventLog interface {
	Append(ctx context.Context, rec events.Record) error
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ProjectDTO struct {
	ID           ids.ProjectID
	Name         string
	Outcome      string
	Status       domain.Status
	TargetValue  *decimal.Decimal
	CurrentValue decimal.Decimal
	Unit         string
	TargetDate   *time.Time
	SphereIDs    []ids.SphereID
	SphereNames  []string
	CreatedAt    time.Time
}

func ToProjectDTO(p domain.Project, sphereNames []string) ProjectDTO {
	unit := ""
	if p.Unit != nil {
		unit = *p.Unit
	}
	outcome := ""
	if p.Outcome != nil {
		outcome = *p.Outcome
	}
	if sphereNames == nil {
		sphereNames = []string{}
	}
	idsCopy := append([]ids.SphereID(nil), p.SphereIDs...)
	return ProjectDTO{
		ID: p.ID, Name: p.Name, Outcome: outcome, Status: p.Status,
		TargetValue: p.TargetValue, CurrentValue: p.CurrentValue, Unit: unit,
		TargetDate: p.TargetDate, SphereIDs: idsCopy, SphereNames: sphereNames, CreatedAt: p.CreatedAt,
	}
}

type CreateProject struct {
	store      ProjectStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewCreateProject(store ProjectStore, events EventLog, transactor Transactor) *CreateProject {
	return &CreateProject{
		store: store, events: events, transactor: transactor,
		now: func() time.Time { return time.Now().UTC() },
	}
}

type CreateProjectInput struct {
	UserID      ids.UserID
	Name        string
	SphereIDs   []ids.SphereID
	Outcome     string
	TargetValue *decimal.Decimal
	Unit        string
	TargetDate  *time.Time
	Source      events.Source
}

func (uc *CreateProject) Execute(ctx context.Context, in CreateProjectInput) (ProjectDTO, error) {
	if in.UserID.IsZero() {
		return ProjectDTO{}, fmt.Errorf("user id is required")
	}
	now := uc.now()
	project, err := domain.NewProject(in.UserID, strings.TrimSpace(in.Name), in.SphereIDs, now)
	if err != nil {
		return ProjectDTO{}, err
	}
	if in.Outcome != "" {
		outcome := strings.TrimSpace(in.Outcome)
		project.Outcome = &outcome
	}
	project.TargetValue = in.TargetValue
	if in.Unit != "" {
		u := strings.TrimSpace(in.Unit)
		project.Unit = &u
	}
	project.TargetDate = in.TargetDate
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.Save(txCtx, project); err != nil {
			return err
		}
		if err := uc.store.SetSpheres(txCtx, project.ID, project.SphereIDs); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        project.UserID,
			AggregateType: "project",
			AggregateID:   project.ID.UUID(),
			EventType:     "ProjectCreated",
			Payload:       map[string]any{"name": project.Name, "sphere_ids": project.SphereIDs},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return ProjectDTO{}, fmt.Errorf("create project: %w", err)
	}
	return ToProjectDTO(project, nil), nil
}

type FindProjectByName struct {
	store ProjectStore
}

func NewFindProjectByName(store ProjectStore) *FindProjectByName {
	return &FindProjectByName{store: store}
}

func (uc *FindProjectByName) Execute(ctx context.Context, userID ids.UserID, name string) (ProjectDTO, error) {
	project, err := uc.store.FindByName(ctx, userID, name)
	if err != nil {
		return ProjectDTO{}, err
	}
	return ToProjectDTO(project, nil), nil
}

type ListProjects struct {
	store ProjectStore
}

func NewListProjects(store ProjectStore) *ListProjects {
	return &ListProjects{store: store}
}

type ListProjectsInput struct {
	UserID   ids.UserID
	SphereID ids.SphereID
}

func (uc *ListProjects) Execute(ctx context.Context, in ListProjectsInput) ([]ProjectDTO, error) {
	if in.UserID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	var items []domain.Project
	var err error
	if !in.SphereID.IsZero() {
		items, err = uc.store.ListBySphere(ctx, in.UserID, in.SphereID)
	} else {
		items, err = uc.store.ListActive(ctx, in.UserID)
	}
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	out := make([]ProjectDTO, 0, len(items))
	for _, item := range items {
		out = append(out, ToProjectDTO(item, nil))
	}
	return out, nil
}
