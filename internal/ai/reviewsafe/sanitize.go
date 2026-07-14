// Package reviewsafe sanitizes assistant review text for Telegram HTML parse mode.
package reviewsafe

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// MaxReviewBodyRunes caps LLM/template body length (header is added by query.Review).
const MaxReviewBodyRunes = 2000

var (
	reBoldOpen  = regexp.MustCompile(`(?i)<b>`)
	reBoldClose = regexp.MustCompile(`(?i)</b>`)
)

const (
	phOpen  = "\x00LIFEOS_B_OPEN\x00"
	phClose = "\x00LIFEOS_B_CLOSE\x00"
)

// SanitizeHTML escapes user/LLM text for Telegram HTML, then restores only exact <b></b>.
// Tags with attributes (e.g. <b onclick=...>) stay escaped. Unbalanced bold is stripped.
func SanitizeHTML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = reBoldOpen.ReplaceAllString(s, phOpen)
	s = reBoldClose.ReplaceAllString(s, phClose)
	s = html.EscapeString(s)
	s = strings.ReplaceAll(s, phOpen, "<b>")
	s = strings.ReplaceAll(s, phClose, "</b>")
	if strings.Count(s, "<b>") != strings.Count(s, "</b>") {
		s = strings.ReplaceAll(s, "<b>", "")
		s = strings.ReplaceAll(s, "</b>", "")
	}
	return TruncateRunes(s, MaxReviewBodyRunes)
}

// EscapePlain escapes text with no allowed markup (task titles, project names, etc.).
func EscapePlain(s string) string {
	return html.EscapeString(strings.TrimSpace(s))
}

// TruncateRunes cuts s to at most max runes without splitting UTF-8.
func TruncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	var b strings.Builder
	b.Grow(max)
	n := 0
	for _, r := range s {
		if n >= max {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}
