package rulebased

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var digitsRE = regexp.MustCompile(`\d`)

// ParseRublesAmount parses Russian money phrases into kopeks (cents).
// Examples: "50 тысяч", "50к", "50000", "50 000 руб".
func ParseRublesAmount(raw string) (int64, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	for _, suffix := range []string{"рублей", "рубля", "руб.", "руб", "₽"} {
		s = strings.ReplaceAll(s, suffix, "")
	}
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")

	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "тысяч"):
		s = strings.TrimSuffix(s, "тысяч")
		mult = 1000
	case strings.HasSuffix(s, "тыс"):
		s = strings.TrimSuffix(s, "тыс")
		mult = 1000
	case strings.HasSuffix(s, "к") && len(s) > 1:
		s = strings.TrimSuffix(s, "к")
		mult = 1000
	}

	s = strings.TrimSpace(s)
	if s == "" || !digitsRE.MatchString(s) {
		return 0, fmt.Errorf("no amount in %q", raw)
	}

	s = strings.ReplaceAll(s, ",", ".")
	if strings.Contains(s, ".") {
		parts := strings.SplitN(s, ".", 2)
		whole, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		frac := parts[1]
		if len(frac) > 2 {
			frac = frac[:2]
		}
		for len(frac) < 2 {
			frac += "0"
		}
		kop, _ := strconv.ParseInt(frac, 10, 64)
		return whole*100 + kop, nil
	}

	whole, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return whole * mult * 100, nil
}

func incomeDescription(text string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "заказ"):
		return "заказ"
	case strings.Contains(lower, "доход"):
		return "доход"
	default:
		return "доход"
	}
}
