package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/projects/domain"
)

type ProjectArchiveStore interface {
	GetByID(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (domain.Project, error)
	FindByName(ctx context.Context, userID ids.UserID, name string) (domain.Project, error)
	UpdateStatus(ctx context.Context, project domain.Project) error
}

type ArchiveProject struct {
	store      ProjectArchiveStore
	events     EventLog
	transactor Transactor
	now        func() time.Time
}

func NewArchiveProject(store ProjectArchiveStore, events EventLog, transactor Transactor) *ArchiveProject {
	return &ArchiveProject{
		store:      store,
		events:     events,
		transactor: transactor,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

type ArchiveProjectInput struct {
	UserID    ids.UserID
	ProjectID ids.ProjectID
	Name      string
	Source    events.Source
}

func (uc *ArchiveProject) Execute(ctx context.Context, in ArchiveProjectInput) (ProjectDTO, error) {
	if in.UserID.IsZero() {
		return ProjectDTO{}, fmt.Errorf("user id is required")
	}

	var project domain.Project
	var err error
	switch {
	case !in.ProjectID.IsZero():
		project, err = uc.store.GetByID(ctx, in.UserID, in.ProjectID)
	case in.Name != "":
		project, err = uc.store.FindByName(ctx, in.UserID, in.Name)
	default:
		return ProjectDTO{}, fmt.Errorf("project id or name is required")
	}
	if errors.Is(err, domain.ErrNotFound) {
		return ProjectDTO{}, err
	}
	if err != nil {
		return ProjectDTO{}, fmt.Errorf("find project: %w", err)
	}

	now := uc.now()
	if err := project.Archive(); err != nil {
		return ProjectDTO{}, err
	}

	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.store.UpdateStatus(txCtx, project); err != nil {
			return err
		}
		return uc.events.Append(txCtx, events.Record{
			UserID:        project.UserID,
			AggregateType: "project",
			AggregateID:   project.ID.UUID(),
			EventType:     "ProjectArchived",
			Payload:       map[string]any{"name": project.Name},
			Source:        in.Source,
			OccurredAt:    now,
		})
	})
	if err != nil {
		return ProjectDTO{}, fmt.Errorf("archive project: %w", err)
	}
	return ProjectDTO{ID: project.ID, Name: project.Name}, nil
}
