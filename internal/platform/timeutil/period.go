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
	startLocal := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, loc)
	endLocal := startLocal.AddDate(0, 0, 7)
	return startLocal.UTC(), endLocal.UTC(), nil
}

func MonthBounds(now time.Time, timezone string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := now.In(loc)
	return MonthBoundsForPeriod(local.Year(), local.Month(), timezone)
}

// MonthBoundsForPeriod returns [start, end) for calendar year/month in timezone.
func MonthBoundsForPeriod(year int, month time.Month, timezone string) (time.Time, time.Time, error) {
	if year < 1 || year > 9999 || month < time.January || month > time.December {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid year-month %04d-%02d", year, month)
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	startLocal := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	endLocal := startLocal.AddDate(0, 1, 0)
	return startLocal.UTC(), endLocal.UTC(), nil
}

func PreviousMonthBounds(now time.Time, timezone string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("load timezone %q: %w", timezone, err)
	}
	local := now.In(loc)
	endLocal := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)
	startLocal := endLocal.AddDate(0, -1, 0)
	return startLocal.UTC(), endLocal.UTC(), nil
}
