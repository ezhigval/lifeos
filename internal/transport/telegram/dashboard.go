package telegram

import (
	"fmt"

	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	taskapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

func FormatDashboard(body string) string {
	if body == "" {
		body = "Выбери раздел reply-кнопками внизу или напиши команду."
	}
	return fmt.Sprintf("<b>LifeOS</b>\n%s\n\n%s", "────────────────", body)
}

func FormatPromptTaskTitle() string {
	return "Введи название задачи:"
}

func FormatPromptProjectTitle() string {
	return "Введи название проекта:"
}

func FormatHomeSummary(taskCount int) string {
	return fmt.Sprintf("🏠 <b>Главная</b>\nЗадач на сегодня: <b>%d</b>\n\nРазделы — reply-кнопки внизу, действия — на экране.", taskCount)
}

func FormatTasksTodayWithActions(items []taskapp.TaskDTO) (string, [][]InlineButton) {
	text := FormatTasksToday(items)
	content := taskDoneButtons(items)
	return text, PrependInline(InlineTasksActions(), content)
}

func taskDoneButtons(items []taskapp.TaskDTO) [][]InlineButton {
	if len(items) == 0 {
		return nil
	}
	var rows [][]InlineButton
	for _, item := range items {
		if item.Status == taskdomain.StatusDone || item.Status == taskdomain.StatusCancelled {
			continue
		}
		rows = append(rows, []InlineButton{
			{Text: "✓ " + truncate(item.Title, 22), CallbackData: CBTaskDone + item.ID.String()},
			{Text: "✕", CallbackData: CBTaskCancel + item.ID.String()},
		})
	}
	return rows
}

func FormatHabitsWithActions(items []habitsapp.HabitDayDTO) (string, [][]InlineButton) {
	text := FormatHabitsToday(items)
	if len(items) == 0 {
		return text, nil
	}
	var rows [][]InlineButton
	for _, item := range items {
		if item.TodayCompleted {
			continue
		}
		rows = append(rows, []InlineButton{{
			Text:         "✓ " + truncate(item.Name, 28),
			CallbackData: CBHabitTrack + item.ID.String(),
		}})
	}
	return text, rows
}

func FormatProjectsPicker(items []projectsapp.ProjectDTO) (string, [][]InlineButton) {
	text := FormatProjects(items)
	content := projectListButtons(items)
	return text, PrependInline(InlineProjectsActions(), content)
}

func projectListButtons(items []projectsapp.ProjectDTO) [][]InlineButton {
	if len(items) == 0 {
		return nil
	}
	var rows [][]InlineButton
	for _, item := range items {
		rows = append(rows, []InlineButton{{
			Text:         truncate(item.Name, 30),
			CallbackData: CBProjectView + item.ID.String(),
		}})
	}
	return rows
}

func FormatProjectTasksWithActions(projectName string, items []taskapp.TaskDTO) (string, [][]InlineButton) {
	text := FormatProjectTasks(projectName, items)
	if len(items) == 0 {
		return text, nil
	}
	var rows [][]InlineButton
	for _, item := range items {
		if item.Status == taskdomain.StatusDone {
			continue
		}
		rows = append(rows, []InlineButton{{
			Text:         "✓ " + truncate(item.Title, 28),
			CallbackData: CBTaskDone + item.ID.String(),
		}})
	}
	return text, rows
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
