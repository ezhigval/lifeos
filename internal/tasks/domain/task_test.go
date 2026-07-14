package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
)

func TestNewTask(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	userID := ids.NewUserID()
	task, err := domain.NewTask(userID, "buy filter", domain.PriorityHigh, nil, now)
	if err != nil {
		t.Fatalf("NewTask() error = %v", err)
	}

	if task.Status != domain.StatusTodo {
		t.Fatalf("status = %s", task.Status)
	}
	if task.Priority != domain.PriorityHigh {
		t.Fatalf("priority = %s", task.Priority)
	}
}

func TestNewTaskRequiresTitle(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTask(ids.NewUserID(), "", domain.PriorityMedium, nil, time.Now())
	if err != domain.ErrEmptyTitle {
		t.Fatalf("error = %v, want ErrEmptyTitle", err)
	}
}

func TestNewTaskInvalidPriority(t *testing.T) {
	t.Parallel()

	_, err := domain.NewTask(ids.NewUserID(), "x", domain.Priority("nope"), nil, time.Now())
	if err != domain.ErrInvalidPriority {
		t.Fatalf("error = %v", err)
	}
}

func TestCompleteTask(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	task, err := domain.NewTask(ids.NewUserID(), "task", domain.PriorityMedium, nil, now)
	if err != nil {
		t.Fatal(err)
	}

	if err := task.Complete(now.Add(time.Hour)); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if task.Status != domain.StatusDone {
		t.Fatalf("status = %s", task.Status)
	}
	if task.CompletedAt == nil {
		t.Fatal("expected completed_at")
	}
}

func TestCannotCompleteCancelledTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task, err := domain.NewTask(ids.NewUserID(), "task", domain.PriorityMedium, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	task.Status = domain.StatusCancelled

	if err := task.Complete(now); err != domain.ErrCannotComplete {
		t.Fatalf("error = %v", err)
	}
}

func TestCannotCompleteAlreadyDoneTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task, err := domain.NewTask(ids.NewUserID(), "task", domain.PriorityMedium, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Complete(now); err != nil {
		t.Fatal(err)
	}
	if err := task.Complete(now); err != domain.ErrAlreadyDone {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateDoneRequiresCompletedAt(t *testing.T) {
	t.Parallel()

	task, err := domain.NewTask(ids.NewUserID(), "task", domain.PriorityMedium, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	task.Status = domain.StatusDone

	if err := task.Validate(); err != domain.ErrCompletedAtNeeded {
		t.Fatalf("error = %v", err)
	}
}

func TestCancelTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task, err := domain.NewTask(ids.NewUserID(), "task", domain.PriorityMedium, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Cancel(now); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if task.Status != domain.StatusCancelled {
		t.Fatalf("status = %s", task.Status)
	}
}

func TestCannotCancelDoneTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task, err := domain.NewTask(ids.NewUserID(), "task", domain.PriorityMedium, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Complete(now); err != nil {
		t.Fatal(err)
	}
	if err := task.Cancel(now); err != domain.ErrCannotCancelDone {
		t.Fatalf("error = %v", err)
	}
}

func TestRescheduleTask(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	task, err := domain.NewTask(ids.NewUserID(), "task", domain.PriorityMedium, &today, now)
	if err != nil {
		t.Fatal(err)
	}
	tomorrow := today.Add(24 * time.Hour)
	if err := task.Reschedule(tomorrow, now); err != nil {
		t.Fatal(err)
	}
	if task.DueDate == nil || !task.DueDate.Equal(tomorrow) {
		t.Fatalf("due = %v", task.DueDate)
	}
}

func TestEditTaskFields(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	task, err := domain.NewTask(ids.NewUserID(), "old", domain.PriorityLow, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	title := "new"
	mins := 45
	tags := []string{"#Work", "home"}
	if err := task.Edit(domain.EditFields{
		Title: &title, DurationMinutes: &mins, Tags: &tags,
	}, now); err != nil {
		t.Fatal(err)
	}
	if task.Title != "new" || task.DurationMinutes == nil || *task.DurationMinutes != 45 {
		t.Fatalf("task = %+v", task)
	}
	found := map[string]bool{}
	for _, tag := range task.Tags {
		found[tag] = true
	}
	if !found["work"] || !found["home"] {
		t.Fatalf("tags = %v", task.Tags)
	}
}

func TestExtractHashtags(t *testing.T) {
	t.Parallel()

	title, tags := domain.ExtractHashtags("купить молоко #шопинг #дом")
	if title != "купить молоко" {
		t.Fatalf("title = %q", title)
	}
	found := map[string]bool{}
	for _, tag := range tags {
		found[tag] = true
	}
	if !found["шопинг"] || !found["дом"] {
		t.Fatalf("tags = %v", tags)
	}
}

func TestApplyEditAndArchiveAndDelete(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	task, err := domain.NewTask(ids.NewUserID(), "task", domain.PriorityMedium, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	due := now.Add(24 * time.Hour)
	desc := "note"
	if err := task.ApplyEdit("new", domain.PriorityHigh, &due, &desc, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if task.Title != "new" || task.Priority != domain.PriorityHigh || task.Description == nil {
		t.Fatalf("edit failed: %+v", task)
	}
	if err := task.Archive(now.Add(2 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	if task.Status != domain.StatusCancelled {
		t.Fatalf("status = %s", task.Status)
	}
	if err := task.SoftDelete(now.Add(3 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	if task.DeletedAt == nil {
		t.Fatal("expected deleted_at")
	}
}
