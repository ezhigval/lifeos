package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

var ErrTaskNotFound = domain.ErrNotFound

type CompleteTaskByTitle struct {
	store    TaskStore
	complete *CompleteTask
}

func NewCompleteTaskByTitle(store TaskStore, complete *CompleteTask) *CompleteTaskByTitle {
	return &CompleteTaskByTitle{store: store, complete: complete}
}

func (uc *CompleteTaskByTitle) Execute(ctx context.Context, userID ids.UserID, title string, source events.Source) (TaskDTO, error) {
	title = strings.TrimSpace(title)
	if userID.IsZero() || title == "" {
		return TaskDTO{}, fmt.Errorf("user id and title are required")
	}

	task, err := uc.store.FindOpenByTitle(ctx, userID, title)
	if err != nil {
		return TaskDTO{}, err
	}

	return uc.complete.Execute(ctx, CompleteTaskInput{
		UserID: userID,
		TaskID: task.ID,
		Source: source,
	})
}
