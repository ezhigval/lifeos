package timeutil

import (
	"fmt"
	"time"
)

func DateInTimezone(now time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC), nil
}

func MonthStartInTimezone(now time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := now.In(loc)
	// UTC instant of local calendar month start (not wall clock forced into UTC).
	startLocal := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)
	return startLocal.UTC(), nil
}
