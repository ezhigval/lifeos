package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type TaskStore interface {
	ListByDueDate(ctx context.Context, userID ids.UserID, dueDate time.Time) ([]taskdomain.Task, error)
	Update(ctx context.Context, task taskdomain.Task) error
}

type TriageOverloadedDay struct {
	store TaskStore
	users UserTimezone
	now   func() time.Time
}

func NewTriageOverloadedDay(store TaskStore, users UserTimezone) *TriageOverloadedDay {
	return &TriageOverloadedDay{
		store: store,
		users: users,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

func (uc *TriageOverloadedDay) Propose(ctx context.Context, userID ids.UserID) (string, []ids.TaskID, error) {
	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	today, err := timeutil.DateInTimezone(uc.now(), tz)
	if err != nil {
		return "", nil, err
	}
	tasks, err := uc.store.ListByDueDate(ctx, userID, today)
	if err != nil {
		return "", nil, err
	}
	var low []ids.TaskID
	var b strings.Builder
	b.WriteString("😮‍💨 <b>Перегруз</b>\n")
	for _, t := range tasks {
		switch t.Priority {
		case taskdomain.PriorityUrgent, taskdomain.PriorityHigh:
			fmt.Fprintf(&b, "🔴 %s\n", t.Title)
		case taskdomain.PriorityMedium:
			fmt.Fprintf(&b, "🟡 %s\n", t.Title)
		default:
			fmt.Fprintf(&b, "🟢 %s (можно перенести)\n", t.Title)
			low = append(low, t.ID)
		}
	}
	if len(low) > 0 {
		b.WriteString("\nПеренести low-priority на завтра?")
	}
	return strings.TrimSpace(b.String()), low, nil
}

func (uc *TriageOverloadedDay) ApplyDefer(ctx context.Context, userID ids.UserID, taskIDs []ids.TaskID) (int, error) {
	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return 0, err
	}
	today, err := timeutil.DateInTimezone(uc.now(), tz)
	if err != nil {
		return 0, err
	}
	tomorrow := today.Add(24 * time.Hour)
	tasks, err := uc.store.ListByDueDate(ctx, userID, today)
	if err != nil {
		return 0, err
	}
	wanted := make(map[ids.TaskID]struct{}, len(taskIDs))
	for _, id := range taskIDs {
		wanted[id] = struct{}{}
	}
	count := 0
	now := uc.now()
	for _, t := range tasks {
		if _, ok := wanted[t.ID]; !ok {
			continue
		}
		if err := t.Reschedule(tomorrow, now); err != nil {
			return count, err
		}
		if err := uc.store.Update(ctx, t); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

type RescheduleTasks struct {
	store TaskStore
	users UserTimezone
	now   func() time.Time
}

func NewRescheduleTasks(store TaskStore, users UserTimezone) *RescheduleTasks {
	return &RescheduleTasks{
		store: store,
		users: users,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

func (uc *RescheduleTasks) Execute(ctx context.Context, userID ids.UserID) (int, error) {
	tz, err := uc.users.Timezone(ctx, userID)
	if err != nil {
		return 0, err
	}
	today, err := timeutil.DateInTimezone(uc.now(), tz)
	if err != nil {
		return 0, err
	}
	tomorrow := today.Add(24 * time.Hour)
	tasks, err := uc.store.ListByDueDate(ctx, userID, today)
	if err != nil {
		return 0, err
	}
	count := 0
	now := uc.now()
	for _, t := range tasks {
		if t.Status == taskdomain.StatusDone || t.Status == taskdomain.StatusCancelled {
			continue
		}
		if err := t.Reschedule(tomorrow, now); err != nil {
			return count, err
		}
		if err := uc.store.Update(ctx, t); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
