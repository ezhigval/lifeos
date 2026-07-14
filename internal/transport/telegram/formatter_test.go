package telegram

import (
	"strings"
	"testing"
	"time"

	taskapp "github.com/valentinezhov/lifeos/internal/tasks/app"
)

func TestFormatReminderScheduledUsesMessageAndLocalAt(t *testing.T) {
	t.Parallel()
	got := FormatReminderScheduled("позвонить", "14.07 19:00")
	if !strings.Contains(got, "позвонить") || !strings.Contains(got, "14.07 19:00") {
		t.Fatalf("got %s", got)
	}
	empty := FormatReminderScheduled("", "14.07 19:00")
	if !strings.Contains(empty, "14.07 19:00") || strings.Contains(empty, "()") {
		t.Fatalf("empty message: %s", empty)
	}
}

func TestEnsureFutureFireAtRollsPast(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 15, 0, 0, 0, time.UTC)
	past := time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC)
	got := ensureFutureFireAt(past, now)
	if !got.After(now) {
		t.Fatalf("expected future, got %s", got)
	}
	if got.Day() != 15 || got.Hour() != 9 {
		t.Fatalf("expected tomorrow 09:00 UTC, got %s", got)
	}
	future := time.Date(2026, 7, 14, 19, 0, 0, 0, time.UTC)
	if keep := ensureFutureFireAt(future, now); !keep.Equal(future) {
		t.Fatalf("future unchanged: %s", keep)
	}
}

func TestFormatTaskCreatedIncludesDurationAndTags(t *testing.T) {
	t.Parallel()
	mins := 45
	got := FormatTaskCreated(taskapp.TaskDTO{
		Title:           "купить <молоко>",
		DurationMinutes: &mins,
		Tags:            []string{"шопинг"},
	})
	if !strings.Contains(got, "45м") || !strings.Contains(got, "#шопинг") {
		t.Fatalf("meta missing: %s", got)
	}
	if !strings.Contains(got, "&lt;молоко&gt;") {
		t.Fatalf("title must be escaped: %s", got)
	}
}

func TestFormatTasksByTagIncludesDurationTagsDue(t *testing.T) {
	t.Parallel()
	mins := 30
	due := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	got := FormatTasksByTag("дом", []taskapp.TaskDTO{{
		Title:           "уборка",
		Priority:        "medium",
		DurationMinutes: &mins,
		Tags:            []string{"дом"},
		DueDate:         &due,
	}})
	if !strings.Contains(got, "30м") || !strings.Contains(got, "#дом") || !strings.Contains(got, "14.07") {
		t.Fatalf("line incomplete: %s", got)
	}
}

func TestFormatTaskRescheduledNilDue(t *testing.T) {
	t.Parallel()
	got := FormatTaskRescheduled(taskapp.TaskDTO{Title: "x"})
	if strings.Contains(got, "→") {
		t.Fatalf("nil due must omit arrow: %s", got)
	}
	due := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	got = FormatTaskRescheduled(taskapp.TaskDTO{Title: "x", DueDate: &due})
	if !strings.Contains(got, "→") || !strings.Contains(got, "15.07") {
		t.Fatalf("due date missing: %s", got)
	}
}

func TestFormatProjectTasksIncludesMeta(t *testing.T) {
	t.Parallel()
	mins := 15
	got := FormatProjectTasks("Свадьба", []taskapp.TaskDTO{{
		Title:           "торт",
		Priority:        "high",
		DurationMinutes: &mins,
		Tags:            []string{"еда"},
	}})
	if !strings.Contains(got, "15м") || !strings.Contains(got, "#еда") {
		t.Fatalf("project task meta missing: %s", got)
	}
}
