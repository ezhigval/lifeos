package app

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/projects/domain"
)

type ProgressDTO struct {
	ProjectID ids.ProjectID
	Name      string
	Target    string
	Current   string
	Remaining string
	Percent   string
	Unit      string
	HasTarget bool
}

type ProjectProgressStore interface {
	GetByID(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (domain.Project, error)
	FindByName(ctx context.Context, userID ids.UserID, name string) (domain.Project, error)
	ListActive(ctx context.Context, userID ids.UserID) ([]domain.Project, error)
}

type GetProjectProgress struct {
	store ProjectProgressStore
}

func NewGetProjectProgress(store ProjectProgressStore) *GetProjectProgress {
	return &GetProjectProgress{store: store}
}

func (uc *GetProjectProgress) Execute(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (ProgressDTO, error) {
	if userID.IsZero() {
		return ProgressDTO{}, fmt.Errorf("user id is required")
	}

	var target domain.Project
	var err error
	switch {
	case !projectID.IsZero():
		target, err = uc.store.GetByID(ctx, userID, projectID)
	default:
		items, listErr := uc.store.ListActive(ctx, userID)
		if listErr != nil {
			return ProgressDTO{}, listErr
		}
		if len(items) == 0 {
			return ProgressDTO{}, fmt.Errorf("no active projects")
		}
		target = items[0]
	}
	if err != nil {
		return ProgressDTO{}, err
	}
	return progressDTOFromProject(target)
}

func (uc *GetProjectProgress) ExecuteByName(ctx context.Context, userID ids.UserID, name string) (ProgressDTO, error) {
	if userID.IsZero() {
		return ProgressDTO{}, fmt.Errorf("user id is required")
	}
	project, err := uc.store.FindByName(ctx, userID, name)
	if err != nil {
		return ProgressDTO{}, err
	}
	return progressDTOFromProject(project)
}

func progressDTOFromProject(project domain.Project) (ProgressDTO, error) {
	progress, err := project.Progress()
	if err != nil {
		return ProgressDTO{
			ProjectID: project.ID,
			Name:      project.Name,
			Current:   project.CurrentValue.String(),
			HasTarget: false,
		}, nil
	}
	return ProgressDTO{
		ProjectID: project.ID,
		Name:      project.Name,
		Target:    progress.Target.String(),
		Current:   progress.Current.String(),
		Remaining: progress.Remaining.String(),
		Percent:   progress.Percent.StringFixed(1),
		Unit:      progress.Unit,
		HasTarget: true,
	}, nil
}

func ParseTarget(s string) (*decimal.Decimal, error) {
	if s == "" {
		return nil, nil
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
