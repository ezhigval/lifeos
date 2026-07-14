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

type fakeStore struct {
	tasks map[ids.TaskID]domain.Task
}

func newFakeStore() *fakeStore {
	return &fakeStore{tasks: make(map[ids.TaskID]domain.Task)}
}

func (s *fakeStore) Save(_ context.Context, task domain.Task) error {
	s.tasks[task.ID] = task
	return nil
}

func (s *fakeStore) GetByID(_ context.Context, userID ids.UserID, taskID ids.TaskID) (domain.Task, error) {
	task, ok := s.tasks[taskID]
	if !ok || task.UserID != userID || task.DeletedAt != nil {
		return domain.Task{}, domain.ErrNotFound
	}
	return task, nil
}

func (s *fakeStore) ListByDueDate(_ context.Context, userID ids.UserID, dueDate time.Time) ([]domain.Task, error) {
	var out []domain.Task
	for _, task := range s.tasks {
		if task.UserID != userID || task.DueDate == nil {
			continue
		}
		if task.DueDate.Format("2006-01-02") == dueDate.Format("2006-01-02") {
			out = append(out, task)
		}
	}
	return out, nil
}

func (s *fakeStore) ListOpenDueOnOrBefore(_ context.Context, userID ids.UserID, dueDate time.Time) ([]domain.Task, error) {
	var out []domain.Task
	for _, task := range s.tasks {
		if task.UserID != userID || task.DueDate == nil {
			continue
		}
		if task.Status != domain.StatusTodo && task.Status != domain.StatusInProgress {
			continue
		}
		if !task.DueDate.After(dueDate) {
			out = append(out, task)
		}
	}
	return out, nil
}

func (s *fakeStore) ListByTag(_ context.Context, userID ids.UserID, tag string) ([]domain.Task, error) {
	var out []domain.Task
	for _, task := range s.tasks {
		if task.UserID != userID {
			continue
		}
		if task.Status != domain.StatusTodo && task.Status != domain.StatusInProgress {
			continue
		}
		for _, t := range task.Tags {
			if t == tag {
				out = append(out, task)
				break
			}
		}
	}
	return out, nil
}

func (s *fakeStore) SetProjects(_ context.Context, taskID ids.TaskID, projectIDs []ids.ProjectID) error {
	task, ok := s.tasks[taskID]
	if !ok {
		return errors.New("not found")
	}
	task.ProjectIDs = projectIDs
	s.tasks[taskID] = task
	return nil
}

func (s *fakeStore) ListByProject(_ context.Context, userID ids.UserID, projectID ids.ProjectID) ([]domain.Task, error) {
	var out []domain.Task
	for _, task := range s.tasks {
		if task.UserID != userID {
			continue
		}
		for _, pid := range task.ProjectIDs {
			if pid == projectID {
				out = append(out, task)
				break
			}
		}
	}
	return out, nil
}

func (s *fakeStore) Update(_ context.Context, task domain.Task) error {
	s.tasks[task.ID] = task
	return nil
}

func (s *fakeStore) FindOpenByTitle(_ context.Context, userID ids.UserID, title string) (domain.Task, error) {
	for _, task := range s.tasks {
		if task.UserID != userID || task.DeletedAt != nil {
			continue
		}
		if task.Status != domain.StatusTodo && task.Status != domain.StatusInProgress {
			continue
		}
		if task.Title == title {
			return task, nil
		}
	}
	return domain.Task{}, domain.ErrNotFound
}

type fakeEvents struct {
	records []events.Record
}

func (f *fakeEvents) Append(_ context.Context, rec events.Record) error {
	f.records = append(f.records, rec)
	return nil
}

type fakeTx struct{}

func (fakeTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeUsers struct {
	tz string
}

func (f fakeUsers) Timezone(context.Context, ids.UserID) (string, error) {
	return f.tz, nil
}

func TestCreateTaskWritesEvent(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	ev := &fakeEvents{}
	uc := app.NewCreateTask(store, ev, fakeTx{}, nil)

	userID := ids.NewUserID()
	dto, err := uc.Execute(context.Background(), app.CreateTaskInput{
		UserID: userID,
		Title:  "buy milk",
		Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if dto.Title != "buy milk" {
		t.Fatalf("title = %q", dto.Title)
	}
	if len(ev.records) != 1 || ev.records[0].EventType != "TaskCreated" {
		t.Fatalf("events = %+v", ev.records)
	}
}

func TestCompleteTaskWritesEvent(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	ev := &fakeEvents{}
	create := app.NewCreateTask(store, ev, fakeTx{}, nil)
	userID := ids.NewUserID()

	created, err := create.Execute(context.Background(), app.CreateTaskInput{
		UserID: userID,
		Title:  "finish sprint",
		Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}

	complete := app.NewCompleteTask(store, ev, fakeTx{})
	done, err := complete.Execute(context.Background(), app.CompleteTaskInput{
		UserID: userID,
		TaskID: created.ID,
		Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if done.Status != domain.StatusDone {
		t.Fatalf("status = %s", done.Status)
	}
	if len(ev.records) != 2 || ev.records[1].EventType != "TaskCompleted" {
		t.Fatalf("events = %+v", ev.records)
	}
}

func TestListTasksTodayUsesTimezone(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	userID := ids.NewUserID()
	// DateInTimezone returns the user's wall-calendar day as UTC midnight.
	// After ~21:00 UTC that day is next calendar date in Europe/Moscow.
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatal(err)
	}
	local := time.Now().UTC().In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)

	task, err := domain.NewTask(userID, "today task", domain.PriorityMedium, &today, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), task); err != nil {
		t.Fatal(err)
	}

	uc := app.NewListTasksToday(store, fakeUsers{tz: "Europe/Moscow"})
	items, err := uc.Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %+v", items)
	}
}

func TestCreateTaskRequiresUserID(t *testing.T) {
	t.Parallel()

	uc := app.NewCreateTask(newFakeStore(), &fakeEvents{}, fakeTx{}, nil)
	_, err := uc.Execute(context.Background(), app.CreateTaskInput{
		Title:  "x",
		Source: events.SourceCLI,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCancelRescheduleAndHashtags(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	ev := &fakeEvents{}
	create := app.NewCreateTask(store, ev, fakeTx{}, nil)
	userID := ids.NewUserID()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	created, err := create.Execute(context.Background(), app.CreateTaskInput{
		UserID: userID, Title: "rollover me #дом", DueDate: &today, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Title != "rollover me" || len(created.Tags) != 1 || created.Tags[0] != "дом" {
		t.Fatalf("created = %+v", created)
	}

	other, err := create.Execute(context.Background(), app.CreateTaskInput{
		UserID: userID, Title: "cancel me", DueDate: &today, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := app.NewCancelTask(store, ev, fakeTx{}).Execute(context.Background(), app.CancelTaskInput{
		UserID: userID, TaskID: other.ID, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Fatalf("status = %s", cancelled.Status)
	}

	tomorrow := today.Add(24 * time.Hour)
	got, err := app.NewRescheduleTask(store, ev, fakeTx{}).Execute(context.Background(), app.RescheduleTaskInput{
		UserID: userID, TaskID: created.ID, DueDate: tomorrow, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.DueDate == nil || got.DueDate.Format("2006-01-02") != "2026-07-15" {
		t.Fatalf("due = %v", got.DueDate)
	}
	if store.tasks[created.ID].DueDate.Format("2006-01-02") != "2026-07-15" {
		t.Fatalf("persisted due not updated")
	}

	tagged, err := app.NewListTasksByTag(store).Execute(context.Background(), userID, "дом")
	if err != nil {
		t.Fatal(err)
	}
	if len(tagged) != 1 || tagged[0].ID != created.ID {
		t.Fatalf("tagged = %+v", tagged)
	}
}

func TestCreateAndEditTaskDescription(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	ev := &fakeEvents{}
	userID := ids.NewUserID()
	desc := " details "
	created, err := app.NewCreateTask(store, ev, fakeTx{}, nil).Execute(context.Background(), app.CreateTaskInput{
		UserID: userID, Title: "with desc", Description: &desc, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Description == nil || *created.Description != "details" {
		t.Fatalf("created desc = %v", created.Description)
	}

	edited, err := app.NewEditTask(store, ev, fakeTx{}, nil).Execute(context.Background(), app.EditTaskInput{
		UserID: userID, TaskID: created.ID, ClearDescription: true, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if edited.Description != nil {
		t.Fatalf("expected cleared description, got %v", edited.Description)
	}
}

func TestCompleteTaskRequiresIDs(t *testing.T) {
	t.Parallel()

	uc := app.NewCompleteTask(newFakeStore(), &fakeEvents{}, fakeTx{})
	_, err := uc.Execute(context.Background(), app.CompleteTaskInput{Source: events.SourceCLI})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListTasksTodayRequiresUserID(t *testing.T) {
	t.Parallel()

	uc := app.NewListTasksToday(newFakeStore(), fakeUsers{tz: "UTC"})
	_, err := uc.Execute(context.Background(), ids.UserID{})
	if err == nil {
		t.Fatal("expected error")
	}
}
