package rulebased

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ParseSleepHours(raw string) (float64, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.TrimSuffix(s, "часов")
	s = strings.TrimSuffix(s, "часа")
	s = strings.TrimSuffix(s, "час")
	s = strings.TrimSuffix(s, "ч")
	s = strings.TrimSuffix(s, "h")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	if s == "" {
		return 0, fmt.Errorf("no sleep duration in %q", raw)
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func SleepHoursToMinutes(hours float64) int32 {
	return int32(math.Round(hours * 60))
}

func HoursMinutesToSleepMinutes(hours, minutes int) int32 {
	return int32(hours*60 + minutes)
}
