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

func TestGetUpdateArchiveDeleteTask(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	ev := &fakeEvents{}
	userID := ids.NewUserID()
	created, err := app.NewCreateTask(store, ev, fakeTx{}, nil).Execute(context.Background(), app.CreateTaskInput{
		UserID: userID, Title: "mutate me", Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := app.NewGetTask(store).Execute(context.Background(), userID, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "mutate me" {
		t.Fatalf("got = %+v", got)
	}
	if _, err := app.NewGetTask(store).Execute(context.Background(), userID, ids.NewTaskID()); !errors.Is(err, app.ErrTaskNotFound) {
		t.Fatalf("missing err = %v", err)
	}

	due := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	desc := "updated"
	updated, err := app.NewUpdateTask(store, ev, fakeTx{}).Execute(context.Background(), app.UpdateTaskInput{
		UserID: userID, TaskID: created.ID, Title: "mutated", Priority: domain.PriorityHigh,
		DueDate: &due, Description: &desc, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "mutated" || updated.Priority != domain.PriorityHigh {
		t.Fatalf("updated = %+v", updated)
	}

	archived, err := app.NewArchiveTask(store, ev, fakeTx{}).Execute(context.Background(), app.ArchiveTaskInput{
		UserID: userID, TaskID: created.ID, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if archived.Status != domain.StatusCancelled {
		t.Fatalf("archived status = %s", archived.Status)
	}

	// recreate open task for delete
	again, err := app.NewCreateTask(store, ev, fakeTx{}, nil).Execute(context.Background(), app.CreateTaskInput{
		UserID: userID, Title: "delete me", Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.NewDeleteTask(store, ev, fakeTx{}).Execute(context.Background(), app.DeleteTaskInput{
		UserID: userID, TaskID: again.ID, Source: events.SourceCLI,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.NewGetTask(store).Execute(context.Background(), userID, again.ID); !errors.Is(err, app.ErrTaskNotFound) {
		t.Fatalf("deleted get err = %v", err)
	}
}

func TestListTasksByProjectAndTitleMutations(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	ev := &fakeEvents{}
	userID := ids.NewUserID()
	projectID := ids.NewProjectID()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	created, err := app.NewCreateTask(store, ev, fakeTx{}, fakeProjects{ok: true}).Execute(context.Background(), app.CreateTaskInput{
		UserID: userID, Title: "proj task", DueDate: &today, ProjectIDs: []ids.ProjectID{projectID}, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}

	listed, err := app.NewListTasksByProject(store).Execute(context.Background(), userID, projectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("listed = %+v", listed)
	}
	if _, err := app.NewListTasksByProject(store).Execute(context.Background(), userID, ids.ProjectID{}); err == nil {
		t.Fatal("expected project id error")
	}

	tomorrow := today.Add(24 * time.Hour)
	rescheduled, err := app.NewRescheduleTaskByTitle(store, app.NewRescheduleTask(store, ev, fakeTx{})).
		Execute(context.Background(), userID, "proj task", tomorrow, events.SourceCLI)
	if err != nil {
		t.Fatal(err)
	}
	if rescheduled.DueDate == nil || rescheduled.DueDate.Format("2006-01-02") != "2026-07-15" {
		t.Fatalf("rescheduled = %+v", rescheduled)
	}

	cancelled, err := app.NewCancelTaskByTitle(store, app.NewCancelTask(store, ev, fakeTx{})).
		Execute(context.Background(), userID, "proj task", events.SourceCLI)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != domain.StatusCancelled {
		t.Fatalf("cancelled = %s", cancelled.Status)
	}

	if _, err := app.NewCancelTaskByTitle(store, app.NewCancelTask(store, ev, fakeTx{})).
		Execute(context.Background(), userID, "missing", events.SourceCLI); !errors.Is(err, app.ErrTaskNotFound) {
		t.Fatalf("cancel missing err = %v", err)
	}
}

func TestAutoRescheduleIncomplete(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	ev := &fakeEvents{}
	userID := ids.NewUserID()
	yesterday := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	task, err := domain.NewTask(userID, "overdue", domain.PriorityMedium, &yesterday, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), task); err != nil {
		t.Fatal(err)
	}

	uc := app.NewAutoRescheduleIncomplete(store, ev, fakeTx{}, fakeUsers{tz: "UTC"})
	// Freeze "now" conceptually: DateInTimezone(UTC now) varies; seed due before today via ListOpenDueOnOrBefore.
	// Execute uses real now — overdue yesterday relative to today should move.
	result, err := uc.Execute(context.Background(), userID, events.SourceCLI)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Moved) != 1 {
		t.Fatalf("moved = %+v", result.Moved)
	}
	if result.Moved[0].DueDate == nil {
		t.Fatal("expected new due date")
	}

	if _, err := uc.Execute(context.Background(), ids.UserID{}, events.SourceCLI); err == nil {
		t.Fatal("expected user id error")
	}
}

type fakeProjects struct {
	ok  bool
	err error
}

func (f fakeProjects) AllExist(context.Context, ids.UserID, []ids.ProjectID) (bool, error) {
	return f.ok, f.err
}

func TestEditTaskFieldsAndProjects(t *testing.T) {
	t.Parallel()

	store := newFakeStore()
	ev := &fakeEvents{}
	userID := ids.NewUserID()
	created, err := app.NewCreateTask(store, ev, fakeTx{}, nil).Execute(context.Background(), app.CreateTaskInput{
		UserID: userID, Title: "edit target", Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}

	title := "edited title"
	prio := domain.PriorityUrgent
	due := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	mins := 45
	tags := []string{"дом", "важное"}
	projectID := ids.NewProjectID()
	pids := []ids.ProjectID{projectID}

	edited, err := app.NewEditTask(store, ev, fakeTx{}, fakeProjects{ok: true}).Execute(context.Background(), app.EditTaskInput{
		UserID: userID, TaskID: created.ID,
		Title: &title, Priority: &prio, DueDate: &due, DurationMinutes: &mins,
		Tags: &tags, ProjectIDs: &pids, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if edited.Title != "edited title" || edited.Priority != domain.PriorityUrgent {
		t.Fatalf("edited = %+v", edited)
	}
	if edited.DurationMinutes == nil || *edited.DurationMinutes != 45 {
		t.Fatalf("duration = %v", edited.DurationMinutes)
	}
	if len(edited.ProjectIDs) != 1 || edited.ProjectIDs[0] != projectID {
		t.Fatalf("projects = %+v", edited.ProjectIDs)
	}

	cleared, err := app.NewEditTask(store, ev, fakeTx{}, nil).Execute(context.Background(), app.EditTaskInput{
		UserID: userID, TaskID: created.ID, ClearDueDate: true, ClearDuration: true, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cleared.DueDate != nil || cleared.DurationMinutes != nil {
		t.Fatalf("cleared = %+v", cleared)
	}

	if _, err := app.NewEditTask(store, ev, fakeTx{}, nil).Execute(context.Background(), app.EditTaskInput{
		Source: events.SourceCLI,
	}); err == nil {
		t.Fatal("expected ids required")
	}
	if _, err := app.NewEditTask(store, ev, fakeTx{}, fakeProjects{ok: false}).Execute(context.Background(), app.EditTaskInput{
		UserID: userID, TaskID: created.ID, ProjectIDs: &pids, Source: events.SourceCLI,
	}); err == nil {
		t.Fatal("expected project not found")
	}
}
<<<<<<< HEAD

=======
>>>>>>> origin/cursor/prod-hosting-fe85
func TestListTasksDueBetween(t *testing.T) {
	t.Parallel()
	store := newFakeStore()
	ev := &fakeEvents{}
	userID := ids.NewUserID()
	from := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	created, err := app.NewCreateTask(store, ev, fakeTx{}, nil).Execute(context.Background(), app.CreateTaskInput{
		UserID: userID, Title: "in range", DueDate: &due, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	listed, err := app.NewListTasksDueBetween(store).Execute(context.Background(), userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("listed=%+v", listed)
	}
	if _, err := app.NewListTasksDueBetween(store).Execute(context.Background(), ids.UserID{}, from, to); err == nil {
		t.Fatal("expected zero user error")
	}
	if _, err := app.NewListTasksDueBetween(store).Execute(context.Background(), userID, to, from); err == nil {
		t.Fatal("expected invalid range")
	}
}
