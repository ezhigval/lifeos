package telegram

import "strings"

const (
	CmdStart    = "/start"
	CmdCancel   = "/cancel"
	CmdClear    = "/clear"
	CmdDelete   = "/delete"
	CmdTriage   = "/triage"
	CmdToday    = "/today"
	CmdTasks    = "/tasks"
	CmdHabits   = "/habits"
	CmdProjects = "/projects"
	CmdCalendar = "/calendar"
	TextCancel  = "отмена"
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
	case CmdTriage:
		return ActionTriage, true
	case CmdToday, CmdTasks:
		return ActionTasksToday, true
	case CmdHabits:
		return ActionHabits, true
	case CmdProjects:
		return ActionProjects, true
	case CmdCalendar:
		return ActionCalendar, true
	default:
		return "", false
	}
}

// BotCommandInfo mirrors the Telegram Bot API "BotCommand" object used by
// setMyCommands to populate the client-side "/" command list.
type BotCommandInfo struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// DefaultBotCommands is the command menu advertised to Telegram clients.
// Keep in sync with commandToAction above; /triage is the MA-C6 entry point.
func DefaultBotCommands() []BotCommandInfo {
	return []BotCommandInfo{
		{Command: "start", Description: "Главная"},
		{Command: "today", Description: "Задачи на сегодня"},
		{Command: "tasks", Description: "Приоритеты (top 3)"},
		{Command: "triage", Description: "Разбор просрочек"},
		{Command: "habits", Description: "Привычки сегодня"},
		{Command: "projects", Description: "Проекты"},
		{Command: "calendar", Description: "Календарь на сегодня"},
		{Command: "clear", Description: "Очистить экран"},
		{Command: "cancel", Description: "Отменить ввод"},
	}
}
