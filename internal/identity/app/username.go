package app

import (
	"strings"
	"unicode"
)

// NormalizeTelegramUsername strips @ and lowercases a public Telegram nick.
// Rules match Telegram: 5–32 chars, [a-z0-9_], must start with a letter.
func NormalizeTelegramUsername(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "@")
	s = strings.ToLower(s)
	if len(s) < 5 || len(s) > 32 {
		return "", false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9' && i > 0:
		case r == '_' && i > 0:
		default:
			if unicode.IsSpace(r) {
				return "", false
			}
			return "", false
		}
	}
	return s, true
}
