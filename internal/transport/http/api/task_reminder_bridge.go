package api

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

// syncTaskReminder keeps scheduled_jobs in sync with task kind=reminder + due_date.
// Date-only due fires at the user's morning_review_at on that calendar day.
func (rt *Router) syncTaskReminder(ctx context.Context, userID ids.UserID, dto tasksapp.TaskDTO) {
	rt.cancelTaskReminders(ctx, userID, dto.ID)
	if dto.Kind != taskdomain.KindReminder || dto.DueDate == nil {
		return
	}
	if rt.deps.ScheduleReminder == nil {
		return
	}

	tz := "Europe/Moscow"
	if rt.deps.GetUserByID != nil {
		if user, err := rt.deps.GetUserByID.Execute(ctx, userID); err == nil && user.Timezone != "" {
			tz = user.Timezone
		}
	}
	todHour, todMin := 9, 0
	if rt.deps.GetSettings != nil {
		if s, err := rt.deps.GetSettings.Execute(ctx, userID); err == nil {
			todHour, todMin = s.MorningReviewAt.Hour, s.MorningReviewAt.Minute
		}
	}

	y, m, d := dto.DueDate.UTC().Date()
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	fireAt := time.Date(y, m, d, todHour, todMin, 0, 0, loc).UTC()
	if !fireAt.After(time.Now().UTC()) {
		// Past slot — skip; user can set a future due date.
		return
	}

	msg := fmt.Sprintf("🔔 %s", dto.Title)
	if _, err := rt.deps.ScheduleReminder.Execute(ctx, notifapp.ScheduleReminderInput{
		UserID:  userID,
		Message: msg,
		FireAt:  fireAt,
		TaskID:  dto.ID.String(),
	}); err != nil {
		rt.log().Warn("schedule task reminder failed", "task_id", dto.ID.String(), "error", err)
	}
}

func (rt *Router) cancelTaskReminders(ctx context.Context, userID ids.UserID, taskID ids.TaskID) {
	if rt.deps.CancelReminder == nil || taskID.IsZero() {
		return
	}
	if err := rt.deps.CancelReminder.CancelForTask(ctx, userID, taskID.String()); err != nil {
		rt.log().Warn("cancel task reminders failed", "task_id", taskID.String(), "error", err)
	}
}

func (rt *Router) log() *slog.Logger {
	if rt.deps.Log != nil {
		return rt.deps.Log
	}
	return slog.Default()
}
