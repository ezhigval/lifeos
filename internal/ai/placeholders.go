package ai

import "strings"

// IsPlaceholderTitle reports invented / stub task titles that must never be persisted.
// Used by agent tools, intent dispatch, and classifiers.
func IsPlaceholderTitle(title string) bool {
	switch strings.ToLower(strings.TrimSpace(title)) {
	case "", "разобрать входящие", "разобрать почту", "задача", "todo", "new task", "новая задача":
		return true
	default:
		return false
	}
}
