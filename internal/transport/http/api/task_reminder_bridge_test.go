package api

import (
	"context"
	"testing"
	"time"

	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type captureScheduler struct {
	calls []notifapp.ScheduleReminderInput
}

func (c *captureScheduler) Execute(_ context.Context, in notifapp.ScheduleReminderInput) (notifapp.ReminderDTO, error) {
	c.calls = append(c.calls, in)
	return notifapp.ReminderDTO{Message: in.Message, FireAt: in.FireAt, Status: "pending"}, nil
}

type captureCanceller struct {
	taskIDs []string
}

func (c *captureCanceller) Execute(context.Context, notifapp.CancelReminderInput) (notifapp.ReminderDTO, error) {
	return notifapp.ReminderDTO{}, nil
}

func (c *captureCanceller) CancelForTask(_ context.Context, _ ids.UserID, taskID string) error {
	c.taskIDs = append(c.taskIDs, taskID)
	return nil
}

func TestSyncTaskReminderSchedulesAtMorningReview(t *testing.T) {
	t.Parallel()

	userID := ids.NewUserID()
	taskID := ids.NewTaskID()
	due := time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC)
	sched := &captureScheduler{}
	cancel := &captureCanceller{}
	rt := NewRouter(Deps{ScheduleReminder: sched, CancelReminder: cancel})

	rt.syncTaskReminder(context.Background(), userID, tasksapp.TaskDTO{
		ID:      taskID,
		Title:   "Позвонить",
		Kind:    taskdomain.KindReminder,
		DueDate: &due,
	})

	if len(cancel.taskIDs) != 1 || cancel.taskIDs[0] != taskID.String() {
		t.Fatalf("cancel calls = %#v", cancel.taskIDs)
	}
	if len(sched.calls) != 1 {
		t.Fatalf("schedule calls = %d", len(sched.calls))
	}
	got := sched.calls[0]
	if got.UserID != userID || got.TaskID != taskID.String() || got.Message != "🔔 Позвонить" {
		t.Fatalf("scheduled = %+v", got)
	}
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2030, 6, 15, 9, 0, 0, 0, loc).UTC()
	if !got.FireAt.Equal(want) {
		t.Fatalf("fireAt = %s, want %s", got.FireAt, want)
	}
}

func TestSyncTaskReminderSkipsNonReminder(t *testing.T) {
	t.Parallel()

	taskID := ids.NewTaskID()
	due := time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC)
	sched := &captureScheduler{}
	cancel := &captureCanceller{}
	rt := NewRouter(Deps{ScheduleReminder: sched, CancelReminder: cancel})

	rt.syncTaskReminder(context.Background(), ids.NewUserID(), tasksapp.TaskDTO{
		ID:      taskID,
		Title:   "Обычная",
		Kind:    taskdomain.KindTask,
		DueDate: &due,
	})

	if len(cancel.taskIDs) != 1 {
		t.Fatalf("cancel calls = %d", len(cancel.taskIDs))
	}
	if len(sched.calls) != 0 {
		t.Fatalf("schedule calls = %d", len(sched.calls))
	}
}
