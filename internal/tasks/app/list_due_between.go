package app

import (
	"context"
	"fmt"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type ListTasksDueBetween struct {
	store TaskStore
}

func NewListTasksDueBetween(store TaskStore) *ListTasksDueBetween {
	return &ListTasksDueBetween{store: store}
}

func (uc *ListTasksDueBetween) Execute(ctx context.Context, userID ids.UserID, from, to time.Time) ([]TaskDTO, error) {
	if userID.IsZero() {
		return nil, fmt.Errorf("user id is required")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("invalid date range")
	}
	items, err := uc.store.ListOpenDueBetween(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	return ToDTOs(items), nil
}
