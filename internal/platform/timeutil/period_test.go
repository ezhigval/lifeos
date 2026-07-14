package timeutil_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
)

func TestMonthBoundsMatchesForPeriodMoscow(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	s, e, err := timeutil.MonthBounds(now, "Europe/Moscow")
	if err != nil {
		t.Fatal(err)
	}
	s2, e2, err := timeutil.MonthBoundsForPeriod(2026, time.July, "Europe/Moscow")
	if err != nil {
		t.Fatal(err)
	}
	if !s.Equal(s2) || !e.Equal(e2) {
		t.Fatalf("MonthBounds=%s..%s ForPeriod=%s..%s", s, e, s2, e2)
	}
	// Moscow is UTC+3: July = [Jun 30 21:00Z, Jul 31 21:00Z)
	wantStart := time.Date(2026, 6, 30, 21, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 7, 31, 21, 0, 0, 0, time.UTC)
	if !s.Equal(wantStart) || !e.Equal(wantEnd) {
		t.Fatalf("got %s..%s want %s..%s", s, e, wantStart, wantEnd)
	}
}

func TestPreviousMonthBoundsMoscow(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	s, e, err := timeutil.PreviousMonthBounds(now, "Europe/Moscow")
	if err != nil {
		t.Fatal(err)
	}
	wantStart := time.Date(2026, 5, 31, 21, 0, 0, 0, time.UTC) // Jun 1 00:00 MSK
	wantEnd := time.Date(2026, 6, 30, 21, 0, 0, 0, time.UTC)   // Jul 1 00:00 MSK
	if !s.Equal(wantStart) || !e.Equal(wantEnd) {
		t.Fatalf("got %s..%s want %s..%s", s, e, wantStart, wantEnd)
	}
}

func TestMonthStartInTimezoneIsLocalMidnightUTC(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	start, err := timeutil.MonthStartInTimezone(now, "Europe/Moscow")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 6, 30, 21, 0, 0, 0, time.UTC)
	if !start.Equal(want) {
		t.Fatalf("got %s want %s", start, want)
	}
}

func TestWeekBoundsMondayLocal(t *testing.T) {
	t.Parallel()
	// Wednesday 2026-07-15 12:00 UTC = 15:00 MSK → week Mon 13 Jul 00:00 MSK
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	s, e, err := timeutil.WeekBounds(now, "Europe/Moscow")
	if err != nil {
		t.Fatal(err)
	}
	wantStart := time.Date(2026, 7, 12, 21, 0, 0, 0, time.UTC) // Jul 13 00:00 MSK
	wantEnd := time.Date(2026, 7, 19, 21, 0, 0, 0, time.UTC)   // Jul 20 00:00 MSK
	if !s.Equal(wantStart) || !e.Equal(wantEnd) {
		t.Fatalf("got %s..%s want %s..%s", s, e, wantStart, wantEnd)
	}
}
