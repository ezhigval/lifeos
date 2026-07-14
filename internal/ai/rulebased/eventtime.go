package rulebased

import (
	"strings"
	"time"
)

func ParseEventStart(now time.Time, timezone, dayHint string, hour, minute int) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	local := now.In(loc)
	dayHint = strings.ToLower(strings.TrimSpace(dayHint))

	var day time.Time
	switch {
	case strings.Contains(dayHint, "завтра"):
		t := local.Add(24 * time.Hour)
		day = time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, loc)
	default:
		day = time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, loc)
	}
	return day.UTC(), nil
}
