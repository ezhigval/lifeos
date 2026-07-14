package domain

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseTimeOfDay(raw string) (TimeOfDay, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return TimeOfDay{}, ErrInvalidReviewTime
	}
	parts := strings.Split(raw, ":")
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return TimeOfDay{}, fmt.Errorf("invalid hour in %q", raw)
	}
	minute := 0
	if len(parts) > 1 {
		minute, err = strconv.Atoi(parts[1])
		if err != nil || minute < 0 || minute > 59 {
			return TimeOfDay{}, fmt.Errorf("invalid minute in %q", raw)
		}
	}
	tod := TimeOfDay{Hour: hour, Minute: minute}
	if !tod.Valid() {
		return TimeOfDay{}, ErrInvalidReviewTime
	}
	return tod, nil
}

func ToPGTime(tod TimeOfDay) string {
	return fmt.Sprintf("%02d:%02d:00", tod.Hour, tod.Minute)
}
