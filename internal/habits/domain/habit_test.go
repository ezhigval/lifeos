package domain

import (
	"testing"
	"time"
)

func TestStreak(t *testing.T) {
	t.Parallel()
	today := date("2026-07-14")
	got := Streak([]time.Time{
		date("2026-07-14"),
		date("2026-07-13"),
		date("2026-07-12"),
		date("2026-07-10"),
	}, today)
	if got != 3 {
		t.Fatalf("got %d want 3", got)
	}
}

func TestStreakWithoutToday(t *testing.T) {
	t.Parallel()
	today := date("2026-07-14")
	got := Streak([]time.Time{
		date("2026-07-13"),
		date("2026-07-12"),
	}, today)
	if got != 2 {
		t.Fatalf("got %d want 2", got)
	}
}

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}
