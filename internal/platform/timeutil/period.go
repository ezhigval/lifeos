package timeutil

import (
	"fmt"
	"time"
)

func WeekBounds(now time.Time, timezone string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := now.In(loc)
	daysFromMonday := (int(local.Weekday()) + 6) % 7
	monday := local.AddDate(0, 0, -daysFromMonday)
	start := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, loc).UTC()
	end := start.AddDate(0, 0, 7)
	return start, end, nil
}

func MonthBounds(now time.Time, timezone string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := now.In(loc)
	start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc).UTC()
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

func PreviousMonthBounds(now time.Time, timezone string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := now.In(loc)
	end := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc).UTC()
	start := end.AddDate(0, -1, 0)
	return start, end, nil
}
