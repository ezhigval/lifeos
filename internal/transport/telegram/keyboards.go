package telegram

// Reply menu — основная навигация (разделы + настройки).
const (
	MenuHome       = "🏠 Главная"
	MenuTasksToday = "📅 Задачи"
	MenuHabits     = "🔄 Привычки"
	MenuProjects   = "📁 Проекты"
	MenuCalendar   = "📆 Календарь"
	MenuAnalytics  = "📊 Статистика"
	MenuSettings   = "⚙️ Настройки"

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
)

// MainReplyKeyboard — постоянная reply-клавиатура: разделы и настройки.
func MainReplyKeyboard() [][]string {
	return [][]string{
		{MenuHome, MenuTasksToday},
		{MenuProjects, MenuHabits},
		{MenuCalendar, MenuAnalytics},
		{MenuSettings},
	}
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
