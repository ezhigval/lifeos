package rulebased

import (
	"strings"
	"time"
)

func ParseFireAt(now time.Time, timezone, hint string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	local := now.In(loc)
	hint = strings.ToLower(strings.TrimSpace(hint))

	switch {
	case strings.Contains(hint, "утром"):
		return time.Date(local.Year(), local.Month(), local.Day(), 9, 0, 0, 0, loc).UTC()
	case strings.Contains(hint, "вечером"):
		return time.Date(local.Year(), local.Month(), local.Day(), 19, 0, 0, 0, loc).UTC()
	case strings.Contains(hint, "завтра"):
		t := local.Add(24 * time.Hour)
		return time.Date(t.Year(), t.Month(), t.Day(), 10, 0, 0, 0, loc).UTC()
	case strings.Contains(hint, "через"):
		return now.Add(2 * time.Minute).UTC()
	default:
		return time.Date(local.Year(), local.Month(), local.Day(), 19, 0, 0, 0, loc).UTC()
	}
}

// EnsureFutureFireAt rolls a parsed local wall-clock time forward by 24h until it is
// strictly after now (e.g. "утром" said in the afternoon → tomorrow 09:00).
func EnsureFutureFireAt(fireAt, now time.Time) time.Time {
	fireAt = fireAt.UTC()
	now = now.UTC()
	for !fireAt.After(now) {
		fireAt = fireAt.Add(24 * time.Hour)
	}
	return fireAt
}

func ParseAvailableUntil(hour, minute int, timezone string, now time.Time) time.Time {
	loc, _ := time.LoadLocation(timezone)
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, loc).UTC()
}
