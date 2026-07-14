package rulebased

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseStepCount(raw string) (int32, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, fmt.Errorf("no steps in %q", raw)
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}
