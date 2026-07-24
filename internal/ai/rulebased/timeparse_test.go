package rulebased_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
)

func TestParseFireAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	got := rulebased.ParseFireAt(now, "Europe/Moscow", "через")
	if got.Sub(now) < time.Minute {
		t.Fatalf("expected future fire at, got %s", got)
	}

	evening := rulebased.ParseFireAt(now, "Europe/Moscow", "вечером")
	if evening.Hour() != 16 { // 19:00 MSK = 16:00 UTC
		t.Fatalf("evening = %s", evening)
	}
}

func TestEnsureFutureFireAt(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 15, 0, 0, 0, time.UTC)
	pastMorning := time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC)
	got := rulebased.EnsureFutureFireAt(pastMorning, now)
	want := pastMorning.Add(24 * time.Hour)
	if !got.Equal(want) {
		t.Fatalf("got %s want %s", got, want)
	}
	future := now.Add(time.Hour)
	if keep := rulebased.EnsureFutureFireAt(future, now); !keep.Equal(future) {
		t.Fatalf("future rolled unexpectedly: %s", keep)
	}
}

func TestParseAvailableUntil(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	got := rulebased.ParseAvailableUntil(15, 30, "Europe/Moscow", now)
	if got.Hour() != 12 || got.Minute() != 30 {
		t.Fatalf("until = %s", got)
	}
}
