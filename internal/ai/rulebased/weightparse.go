package rulebased

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseWeightKg(raw string) (float64, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.TrimSuffix(s, "кг")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	if s == "" {
		return 0, fmt.Errorf("no weight in %q", raw)
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}
