package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/app"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

func TestCompleteTaskByTitle(t *testing.T) {
	t.Parallel()

	userID := ids.NewUserID()
	taskID := ids.NewTaskID()
	store := &titleLookupStore{
		task: domain.Task{
			ID:     taskID,
			UserID: userID,
			Title:  "купить фильтр",
			Status: domain.StatusTodo,
		},
	}
	complete := app.NewCompleteTask(store, &fakeEvents{}, fakeTx{})
	uc := app.NewCompleteTaskByTitle(store, complete)

	got, err := uc.Execute(context.Background(), userID, "фильтр", events.SourceCLI)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.StatusDone {
		t.Fatalf("status = %s", got.Status)
	}
}

func TestCompleteTaskByTitleNotFound(t *testing.T) {
	t.Parallel()

	store := &titleLookupStore{err: app.ErrTaskNotFound}
	uc := app.NewCompleteTaskByTitle(store, app.NewCompleteTask(store, &fakeEvents{}, fakeTx{}))

	_, err := uc.Execute(context.Background(), ids.NewUserID(), "missing", events.SourceCLI)
	if !errors.Is(err, app.ErrTaskNotFound) {
		t.Fatalf("err = %v", err)
	}
}

type titleLookupStore struct {
	task domain.Task
	err  error
}

func (s *titleLookupStore) FindOpenByTitle(_ context.Context, _ ids.UserID, _ string) (domain.Task, error) {
	if s.err != nil {
		return domain.Task{}, s.err
	}
	return s.task, nil
}

func (s *titleLookupStore) Save(_ context.Context, task domain.Task) error {
	s.task = task
	return nil
}

func (s *titleLookupStore) GetByID(_ context.Context, userID ids.UserID, taskID ids.TaskID) (domain.Task, error) {
	if s.task.ID == taskID && s.task.UserID == userID {
		return s.task, nil
	}
	return domain.Task{}, errors.New("not found")
}

func (s *titleLookupStore) ListByDueDate(context.Context, ids.UserID, time.Time) ([]domain.Task, error) {
	return nil, nil
}

func (s *titleLookupStore) ListOpenDueOnOrBefore(context.Context, ids.UserID, time.Time) ([]domain.Task, error) {
	return nil, nil
}

func (s *titleLookupStore) ListOpenDueBetween(context.Context, ids.UserID, time.Time, time.Time) ([]domain.Task, error) {
	return nil, nil
}

func (s *titleLookupStore) ListByTag(context.Context, ids.UserID, string) ([]domain.Task, error) {
	return nil, nil
}

func (s *titleLookupStore) SetProjects(context.Context, ids.TaskID, []ids.ProjectID) error {
	return nil
}

func (s *titleLookupStore) ListByProject(context.Context, ids.UserID, ids.ProjectID) ([]domain.Task, error) {
	return nil, nil
}

func (s *titleLookupStore) Update(_ context.Context, task domain.Task) error {
	s.task = task
	return nil
}
