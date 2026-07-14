package telegram

import "strings"

// Reply menu — основная навигация (разделы + настройки + Mini App).
const (
	MenuHome       = "🏠 Главная"
	MenuTasksToday = "📅 Задачи"
	MenuHabits     = "🔄 Привычки"
	MenuProjects   = "📁 Проекты"
	MenuCalendar   = "📆 Календарь"
	MenuAnalytics  = "📊 Статистика"
	MenuSettings   = "⚙️ Настройки"
	MenuMiniApp    = "📱 Mini App"

	// Ситуативные действия (только inline).
	MenuAddTask         = "➕ Задача"
	MenuPriorities      = "🔥 Важное"
	MenuTriage          = "📦 Завал"
	MenuAddProject      = "➕ Проект"
	MenuProjectProgress = "🎯 Прогресс"
)

// Callback data prefixes.
const (
	CBActionPrefix = "action:"
	CBTriageDefer  = "triage:defer"
	CBTaskDone     = "task:done:"
	CBTaskCancel   = "task:cancel:"
	CBHabitTrack   = "habit:track:"
	CBProjectView  = "project:view:"
)

const (
	ActionHome            = "home"
	ActionTasksToday      = "tasks_today"
	ActionPriorities      = "priorities"
	ActionAddTask         = "add_task"
	ActionProjectProgress = "project_progress"
	ActionAddProject      = "add_project"
	ActionTriage          = "triage"
	ActionHabits          = "habits"
	ActionProjects        = "projects"
	ActionCalendar        = "calendar"
	ActionAnalytics       = "analytics"
	ActionSettings        = "settings"
	ActionMiniApp         = "miniapp"
)

// ReplyButton is a Telegram reply-keyboard button (plain text or web_app).
type ReplyButton struct {
	Text   string
	WebApp string // when set, Telegram opens this HTTPS Mini App URL
}

// MainReplyKeyboard — постоянная reply-клавиатура: разделы, настройки и Mini App.
// Mini App на reply-клавиатуре — обычный текст (без web_app): на многих клиентах
// reply web_app открывает WebView с пустым initData. Реальный launch — inline web_app
// (см. InlineOpenMiniApp) или Menu Button.
func MainReplyKeyboard(miniAppURL string) [][]ReplyButton {
	rows := [][]ReplyButton{
		{{Text: MenuHome}, {Text: MenuTasksToday}},
		{{Text: MenuProjects}, {Text: MenuHabits}},
		{{Text: MenuCalendar}, {Text: MenuAnalytics}},
		{{Text: MenuSettings}},
	}
	if strings.TrimSpace(miniAppURL) != "" {
		rows = append(rows, []ReplyButton{{Text: MenuMiniApp}})
	}
	return rows
}

// InlineOpenMiniApp — единственный надёжный launch Mini App из чата (initData OK).
func InlineOpenMiniApp(miniAppURL string) [][]InlineButton {
	url := strings.TrimSpace(miniAppURL)
	if url == "" {
		return nil
	}
	return [][]InlineButton{{{Text: "📱 Открыть Mini App", WebApp: url}}}
}

// InlineHomeActions — быстрые действия на главной.
func InlineHomeActions() [][]InlineButton {
	return [][]InlineButton{
		{
			{Text: MenuAddTask, CallbackData: CBActionPrefix + ActionAddTask},
			{Text: MenuPriorities, CallbackData: CBActionPrefix + ActionPriorities},
		},
		{{Text: MenuTriage, CallbackData: CBActionPrefix + ActionTriage}},
	}
}

// InlineTasksActions — действия в разделе задач.
func InlineTasksActions() [][]InlineButton {
	return [][]InlineButton{
		{
			{Text: MenuAddTask, CallbackData: CBActionPrefix + ActionAddTask},
			{Text: MenuPriorities, CallbackData: CBActionPrefix + ActionPriorities},
		},
		{{Text: MenuTriage, CallbackData: CBActionPrefix + ActionTriage}},
	}
}

// InlineProjectsActions — действия в разделе проектов.
func InlineProjectsActions() [][]InlineButton {
	return [][]InlineButton{
		{
			{Text: MenuAddProject, CallbackData: CBActionPrefix + ActionAddProject},
			{Text: MenuProjectProgress, CallbackData: CBActionPrefix + ActionProjectProgress},
		},
	}
}

// PrependInline ставит строки действий над контекстным списком.
func PrependInline(actions, content [][]InlineButton) [][]InlineButton {
	if len(actions) == 0 {
		return content
	}
	if len(content) == 0 {
		return actions
	}
	out := make([][]InlineButton, 0, len(actions)+len(content))
	out = append(out, actions...)
	out = append(out, content...)
	return out
}
