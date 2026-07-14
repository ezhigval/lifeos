package telegram

import "strings"

const (
	CmdStart  = "/start"
	CmdCancel = "/cancel"
	TextCancel = "отмена"
)

func normalizeBotCommand(text string) string {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") {
		return ""
	}
	if i := strings.Index(text, "@"); i > 0 {
		text = text[:i]
	}
	if i := strings.Index(text, " "); i > 0 {
		text = text[:i]
	}
	return strings.ToLower(text)
}

func isCancelText(text string) bool {
	t := strings.TrimSpace(strings.ToLower(text))
	return t == TextCancel || t == CmdCancel
}

func commandToAction(cmd string) (string, bool) {
	switch cmd {
	case CmdStart:
		return ActionHome, true
	default:
		return "", false
	}
}
