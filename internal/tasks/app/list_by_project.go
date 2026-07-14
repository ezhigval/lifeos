package app

import (
	"context"
	"fmt"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type ListTasksByProject struct {
	store TaskStore
}

func NewListTasksByProject(store TaskStore) *ListTasksByProject {
	return &ListTasksByProject{store: store}
}

func (uc *ListTasksByProject) Execute(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) ([]TaskDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	if projectID.IsZero() {
		return nil, fmt.Errorf("project id is required")
	}
	tasks, err := uc.store.ListByProject(ctx, userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("list tasks by project: %w", err)
	}
	return ToDTOs(tasks), nil
}
