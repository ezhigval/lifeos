package app

import (
	"context"
	"fmt"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type ListTasksByTag struct {
	store TaskStore
}

func NewListTasksByTag(store TaskStore) *ListTasksByTag {
	return &ListTasksByTag{store: store}
}

func (uc *ListTasksByTag) Execute(ctx context.Context, userID ids.UserID, tag string) ([]TaskDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	normalized := domain.NormalizeTags([]string{tag})
	if len(normalized) == 0 {
		return nil, fmt.Errorf("tag is required")
	}
	tasks, err := uc.store.ListByTag(ctx, userID, normalized[0])
	if err != nil {
		return nil, err
	}
	return ToDTOs(tasks), nil
}
