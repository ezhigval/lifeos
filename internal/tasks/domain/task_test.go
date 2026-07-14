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

