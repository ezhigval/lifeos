package telegram

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	calendarapp "github.com/valentinezhov/lifeos/internal/calendar/app"
	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	financedomain "github.com/valentinezhov/lifeos/internal/finance/domain"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	"github.com/valentinezhov/lifeos/internal/query"
	settingsdomain "github.com/valentinezhov/lifeos/internal/settings/domain"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	taskapp "github.com/valentinezhov/lifeos/internal/tasks/app"
)

// Room under Telegram's 4096 cap after FormatDashboard wraps the body.
const maxAssistantHTMLRunes = 3500

// Allowlisted Telegram HTML tags restored after full escape (LLM/review/triage text).
var reTelegramAllowedTag = regexp.MustCompile(`(?i)&lt;(/?)(b|strong|i|em)&gt;`)

// FormatAssistantHTML escapes free-form assistant/review text for parse_mode=HTML,
// restores a small allowlist of emphasis tags, and truncates to a safe length.
// Do not use on bodies already passed through html.EscapeString (double-escape).
func FormatAssistantHTML(text string) string {
	text = sanitizeTelegramHTML(strings.TrimSpace(text))
	r := []rune(text)
	if len(r) <= maxAssistantHTMLRunes {
		return text
	}
	return string(r[:maxAssistantHTMLRunes-1]) + "…"
}

func sanitizeTelegramHTML(s string) string {
	escaped := html.EscapeString(s)
	return reTelegramAllowedTag.ReplaceAllStringFunc(escaped, func(match string) string {
		parts := reTelegramAllowedTag.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		return "<" + parts[1] + strings.ToLower(parts[2]) + ">"
	})
}

func FormatTaskCreated(dto taskapp.TaskDTO) string {
	out := fmt.Sprintf("✅ Задача создана: <b>%s</b>", html.EscapeString(dto.Title))
	if extra := formatTaskMeta(dto.DurationMinutes, dto.Tags); extra != "" {
		out += " · " + html.EscapeString(extra)
	}
	return out
}

func FormatTasksToday(items []taskapp.TaskDTO) string {
	if len(items) == 0 {
		return "Сегодня задач нет."
	}
	var b strings.Builder
	b.WriteString("📅 <b>Задачи на сегодня</b>\n")
	for _, item := range items {
		extra := formatTaskMeta(item.DurationMinutes, item.Tags)
		if extra != "" {
			extra = " · " + extra
		}
		fmt.Fprintf(&b, "• [%s] %s%s\n", item.Priority, html.EscapeString(item.Title), html.EscapeString(extra))
	}
	return strings.TrimSpace(b.String())
}

func formatTaskTags(tags []string) string {
	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		parts = append(parts, "#"+tag)
	}
	return strings.Join(parts, " ")
}

// formatTaskMeta joins duration + hashtags for list/created lines (plain text; escape at call site).
func formatTaskMeta(durationMinutes *int, tags []string) string {
	var parts []string
	if durationMinutes != nil {
		parts = append(parts, fmt.Sprintf("%dм", *durationMinutes))
	}
	if len(tags) > 0 {
		parts = append(parts, formatTaskTags(tags))
	}
	return strings.Join(parts, " · ")
}

func FormatTaskCancelled(dto taskapp.TaskDTO) string {
	return fmt.Sprintf("🚫 Задача отменена: <b>%s</b>", html.EscapeString(dto.Title))
}

func FormatTaskRescheduled(dto taskapp.TaskDTO) string {
	if dto.DueDate == nil {
		return fmt.Sprintf("↪️ Задача перенесена: <b>%s</b>", html.EscapeString(dto.Title))
	}
	due := dto.DueDate.Format("02.01")
	return fmt.Sprintf("↪️ Задача перенесена: <b>%s</b> → <b>%s</b>", html.EscapeString(dto.Title), html.EscapeString(due))
}

func FormatTasksByTag(tag string, items []taskapp.TaskDTO) string {
	if len(items) == 0 {
		return fmt.Sprintf("Открытых задач с тегом <b>#%s</b> нет.", html.EscapeString(tag))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "🏷 <b>Задачи #%s</b>\n", html.EscapeString(tag))
	for _, item := range items {
		extra := formatTaskMeta(item.DurationMinutes, item.Tags)
		if item.DueDate != nil {
			due := item.DueDate.Format("02.01")
			if extra != "" {
				extra = due + " · " + extra
			} else {
				extra = due
			}
		}
		if extra != "" {
			extra = " · " + extra
		}
		fmt.Fprintf(&b, "• [%s] %s%s\n", item.Priority, html.EscapeString(item.Title), html.EscapeString(extra))
	}
	return strings.TrimSpace(b.String())
}

func FormatProjectProgress(p projectsapp.ProgressDTO) string {
	if !p.HasTarget {
		return fmt.Sprintf("Проект <b>%s</b>: текущее значение %s", html.EscapeString(p.Name), p.Current)
	}
	unit := p.Unit
	if unit != "" {
		unit = " " + unit
	}
	return fmt.Sprintf(
		"📁 <b>%s</b>\nЦель: %s%s\nСейчас: %s%s\nОсталось: %s%s\nПрогресс: %s%%",
		html.EscapeString(p.Name), p.Target, unit, p.Current, unit, p.Remaining, unit, p.Percent,
	)
}

func FormatPriorities(items []query.PriorityItem) string {
	if len(items) == 0 {
		return "Приоритетов нет — отличный момент добавить задачу."
	}
	var b strings.Builder
	b.WriteString("🔥 <b>Сейчас важно</b>\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, item.Kind, html.EscapeString(item.Title))
	}
	return strings.TrimSpace(b.String())
}

func FormatReminderScheduled(message, at string) string {
	if strings.TrimSpace(message) == "" {
		return fmt.Sprintf("⏰ Напомню: <b>%s</b>", html.EscapeString(at))
	}
	return fmt.Sprintf("⏰ Напомню: <b>%s</b> (<b>%s</b>)", html.EscapeString(message), html.EscapeString(at))
}

func FormatReminderCancelled(message, at string) string {
	if message == "" {
		return fmt.Sprintf("🚫 Напоминание отменено (<b>%s</b>)", html.EscapeString(at))
	}
	return fmt.Sprintf("🚫 Напоминание отменено: <b>%s</b> (<b>%s</b>)", html.EscapeString(message), html.EscapeString(at))
}

func FormatReminderNotFound(hint string) string {
	if hint == "" {
		return "Не нашёл активное напоминание."
	}
	return fmt.Sprintf("Не нашёл напоминание «%s».", html.EscapeString(hint))
}

func FormatAvailability(until string) string {
	return fmt.Sprintf("🕒 Сегодня работаю до <b>%s</b>", html.EscapeString(until))
}

func FormatTriage(proposal string) string {
	return FormatAssistantHTML(proposal)
}

func FormatTaskCompleted(dto taskapp.TaskDTO) string {
	return fmt.Sprintf("✅ Задача выполнена: <b>%s</b>", html.EscapeString(dto.Title))
}

func FormatTaskNotFound(title string) string {
	return fmt.Sprintf("Не нашёл открытую задачу «%s».", html.EscapeString(title))
}

func FormatMorningReviewSet(at settingsdomain.TimeOfDay) string {
	return fmt.Sprintf("☀️ Утренний обзор в <b>%02d:%02d</b>", at.Hour, at.Minute)
}

func FormatEveningReviewSet(at settingsdomain.TimeOfDay) string {
	return fmt.Sprintf("🌙 Вечерний обзор в <b>%02d:%02d</b>", at.Hour, at.Minute)
}

func FormatQuietHoursSet(sh, sm, eh, em int) string {
	return fmt.Sprintf("🔕 Тихие часы: <b>%02d:%02d — %02d:%02d</b>", sh, sm, eh, em)
}

func FormatIncomeRecorded(dto financeapp.TransactionDTO) string {
	amount := financedomain.FormatMoney(financedomain.Money{
		AmountCents: dto.AmountCents,
		Currency:    dto.Currency,
	})
	month := financedomain.FormatMoney(financedomain.Money{
		AmountCents: dto.MonthTotal,
		Currency:    dto.Currency,
	})
	return fmt.Sprintf(
		"💰 Доход записан: <b>%s</b> (%s)\nЗа месяц: <b>%s</b>",
		amount, html.EscapeString(dto.Description), month,
	)
}

func FormatExpenseRecorded(dto financeapp.TransactionDTO) string {
	amount := financedomain.FormatMoney(financedomain.Money{AmountCents: dto.AmountCents, Currency: dto.Currency})
	month := financedomain.FormatMoney(financedomain.Money{AmountCents: dto.MonthTotal, Currency: dto.Currency})
	return fmt.Sprintf(
		"💸 Расход записан: <b>%s</b> (%s)\nРасходы за месяц: <b>%s</b>",
		amount, html.EscapeString(dto.Description), month,
	)
}

func FormatDebts(items []financeapp.DebtDTO) string {
	if len(items) == 0 {
		return "📋 Открытых долгов нет."
	}
	var b strings.Builder
	b.WriteString("📋 <b>Долги</b>\n")
	for _, item := range items {
		remaining := financedomain.FormatMoney(financedomain.Money{AmountCents: item.RemainingCents, Currency: item.Currency})
		fmt.Fprintf(&b, "• %s — <b>%s</b>\n", html.EscapeString(item.Creditor), remaining)
	}
	return strings.TrimSpace(b.String())
}

func FormatDebtCreated(dto financeapp.DebtDTO) string {
	amount := financedomain.FormatMoney(financedomain.Money{AmountCents: dto.AmountCents, Currency: dto.Currency})
	return fmt.Sprintf("📌 Долг добавлен: <b>%s</b> — %s", html.EscapeString(dto.Creditor), amount)
}

func FormatDebtPaid(dto financeapp.DebtDTO, paidCents int64) string {
	paid := financedomain.FormatMoney(financedomain.Money{AmountCents: paidCents, Currency: dto.Currency})
	if dto.RemainingCents == 0 {
		return fmt.Sprintf("✅ Долг <b>%s</b> погашен — оплачено %s", html.EscapeString(dto.Creditor), paid)
	}
	remaining := financedomain.FormatMoney(financedomain.Money{AmountCents: dto.RemainingCents, Currency: dto.Currency})
	return fmt.Sprintf(
		"💰 Оплата по долгу <b>%s</b>: %s\nОсталось: <b>%s</b>",
		html.EscapeString(dto.Creditor), paid, remaining,
	)
}

func FormatCashFlow(summary financeapp.CashFlowDTO) string {
	income := financedomain.FormatMoney(financedomain.Money{AmountCents: summary.IncomeCents, Currency: summary.Currency})
	expense := financedomain.FormatMoney(financedomain.Money{AmountCents: summary.ExpenseCents, Currency: summary.Currency})
	net := financedomain.FormatMoney(financedomain.Money{AmountCents: summary.NetCents, Currency: summary.Currency})
	return fmt.Sprintf(
		"📊 <b>Финансы за месяц</b>\nДоход: <b>%s</b>\nРасход: <b>%s</b>\nИтого: <b>%s</b>",
		income, expense, net,
	)
}

func FormatHabitCreated(dto habitsapp.HabitDTO) string {
	return fmt.Sprintf("🔄 Привычка добавлена: <b>%s</b>", html.EscapeString(dto.Name))
}

func FormatHabitTracked(result habitsapp.TrackHabitResult) string {
	return fmt.Sprintf("✅ <b>%s</b> — отмечено! Streak: <b>%d</b> 🔥", html.EscapeString(result.Name), result.Streak)
}

func FormatHabitNotFound(name string) string {
	return fmt.Sprintf("Не нашёл привычку «%s». Добавь: «добавь привычку %s»", html.EscapeString(name), html.EscapeString(name))
}

func FormatHabitsToday(items []habitsapp.HabitDayDTO) string {
	if len(items) == 0 {
		return "🔄 Привычек пока нет. Добавь: «добавь привычку бег»"
	}
	var b strings.Builder
	b.WriteString("🔄 <b>Привычки сегодня</b>\n")
	for _, item := range items {
		mark := "○"
		if item.TodayCompleted {
			mark = "✓"
		}
		fmt.Fprintf(&b, "%s %s — streak %d\n", mark, html.EscapeString(item.Name), item.Streak)
	}
	return strings.TrimSpace(b.String())
}

func FormatProjectCreated(dto projectsapp.ProjectDTO) string {
	return fmt.Sprintf("📁 Проект создан: <b>%s</b>", html.EscapeString(dto.Name))
}

func FormatProjectArchived(dto projectsapp.ProjectDTO) string {
	return fmt.Sprintf("📦 Проект <b>%s</b> архивирован", html.EscapeString(dto.Name))
}

func FormatProjectNotFound(name string) string {
	return fmt.Sprintf("Не нашёл проект «%s».", html.EscapeString(name))
}

func FormatProjects(items []projectsapp.ProjectDTO) string {
	if len(items) == 0 {
		return "📁 Проектов пока нет. Добавь: «добавь проект свадьба в сфере Личная жизнь»"
	}
	var b strings.Builder
	b.WriteString("📁 <b>Проекты</b>\nВыбери проект, чтобы увидеть задачи:\n")
	for _, item := range items {
		fmt.Fprintf(&b, "• %s\n", html.EscapeString(item.Name))
	}
	return strings.TrimSpace(b.String())
}

func FormatProjectTasks(projectName string, items []taskapp.TaskDTO) string {
	if len(items) == 0 {
		return fmt.Sprintf("📁 <b>%s</b>\nОткрытых задач нет.", html.EscapeString(projectName))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "📁 <b>%s</b>\n", html.EscapeString(projectName))
	for _, item := range items {
		extra := formatTaskMeta(item.DurationMinutes, item.Tags)
		if extra != "" {
			extra = " · " + extra
		}
		fmt.Fprintf(&b, "• [%s] %s%s\n", item.Priority, html.EscapeString(item.Title), html.EscapeString(extra))
	}
	return strings.TrimSpace(b.String())
}

func FormatCalendarEventCreated(dto calendarapp.EventDTO, timezone string) string {
	when := formatLocalTime(dto.StartsAt, timezone)
	return fmt.Sprintf("📆 Встреча добавлена: <b>%s</b> — %s", html.EscapeString(dto.Title), when)
}

func FormatCalendarToday(items []calendarapp.EventDTO, timezone string) string {
	if len(items) == 0 {
		return "📆 <b>Календарь на сегодня</b>\nСобытий нет."
	}
	var b strings.Builder
	b.WriteString("📆 <b>Календарь на сегодня</b>\n")
	for _, item := range items {
		fmt.Fprintf(&b, "• %s — %s\n", formatLocalTime(item.StartsAt, timezone), html.EscapeString(item.Title))
	}
	return strings.TrimSpace(b.String())
}

func formatLocalTime(t time.Time, timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return t.In(loc).Format("02.01 15:04")
}

func FormatFallback() string {
	return "Не понял. Выбери раздел кнопками внизу или напиши команду."
}

func FormatNoteCreated(dto knowledgeapp.NoteDTO) string {
	preview := truncateNote(dto.Body, 120)
	out := fmt.Sprintf("📝 Заметка сохранена:\n<b>%s</b>", html.EscapeString(preview))
	if len(dto.Tags) > 0 {
		out += "\n" + html.EscapeString(formatNoteTags(dto.Tags))
	}
	return out
}

func FormatNoteDeleted(dto knowledgeapp.NoteDTO) string {
	preview := truncateNote(dto.Body, 80)
	return fmt.Sprintf("🗑 Заметка удалена:\n<b>%s</b>", html.EscapeString(preview))
}

func FormatNoteNotFound(hint string) string {
	if hint == "" {
		return "Не нашёл заметку для удаления."
	}
	return fmt.Sprintf("Не нашёл заметку «%s».", html.EscapeString(hint))
}

func FormatNotes(items []knowledgeapp.NoteDTO) string {
	if len(items) == 0 {
		return "📝 Заметок пока нет."
	}
	var b strings.Builder
	b.WriteString("📝 <b>Последние заметки</b>\n")
	for i, item := range items {
		line := truncateNote(item.Body, 80)
		if len(item.Tags) > 0 {
			line += " " + formatNoteTags(item.Tags)
		}
		fmt.Fprintf(&b, "%d. %s\n", i+1, html.EscapeString(line))
	}
	return strings.TrimSpace(b.String())
}

func formatContactLine(dto careerapp.ContactDTO) string {
	line := dto.Name
	if dto.Company != "" {
		line += " — " + dto.Company
	}
	if dto.Role != "" {
		line += " (" + dto.Role + ")"
	}
	return line
}

func FormatContactCreated(dto careerapp.ContactDTO) string {
	return fmt.Sprintf("👤 Контакт сохранён:\n<b>%s</b>", html.EscapeString(formatContactLine(dto)))
}

func FormatContactDeleted(dto careerapp.ContactDTO) string {
	return fmt.Sprintf("🗑 Контакт удалён:\n<b>%s</b>", html.EscapeString(formatContactLine(dto)))
}

func FormatContactNotFound(hint string) string {
	if hint == "" {
		return "Не нашёл контакт для удаления."
	}
	return fmt.Sprintf("Не нашёл контакт «%s».", html.EscapeString(hint))
}

func FormatContacts(items []careerapp.ContactDTO) string {
	if len(items) == 0 {
		return "👤 Контактов пока нет."
	}
	var b strings.Builder
	b.WriteString("👤 <b>Контакты</b>\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, html.EscapeString(formatContactLine(item)))
	}
	return strings.TrimSpace(b.String())
}

func FormatContactSearchResults(query string, items []careerapp.ContactDTO) string {
	if len(items) == 0 {
		return fmt.Sprintf("По запросу «%s» контактов не найдено.", html.EscapeString(query))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "🔎 <b>Контакты: %s</b>\n", html.EscapeString(query))
	for i, item := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, html.EscapeString(formatContactLine(item)))
	}
	return strings.TrimSpace(b.String())
}

func FormatSphereCreated(dto spheresapp.SphereDTO) string {
	return fmt.Sprintf("🌐 Сфера добавлена: <b>%s</b>", html.EscapeString(dto.Name))
}

func FormatSphereUpdated(dto spheresapp.SphereDTO) string {
	return fmt.Sprintf("🌐 Сфера обновлена: <b>%s</b>", html.EscapeString(dto.Name))
}

func FormatSphereDeleted(dto spheresapp.SphereDTO) string {
	return fmt.Sprintf("🗑 Сфера удалена: <b>%s</b>", html.EscapeString(dto.Name))
}

func FormatSphereNotFound(hint string) string {
	if hint == "" {
		return "Не нашёл сферу для удаления."
	}
	return fmt.Sprintf("Не нашёл сферу «%s».", html.EscapeString(hint))
}

func FormatSpheres(items []spheresapp.SphereDTO) string {
	if len(items) == 0 {
		return "🌐 Сфер пока нет."
	}
	var b strings.Builder
	b.WriteString("🌐 <b>Сферы жизни</b>\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, html.EscapeString(item.Name))
	}
	return strings.TrimSpace(b.String())
}

func formatSkillLine(dto careerapp.SkillDTO) string {
	if dto.Level == "" {
		return dto.Name
	}
	return dto.Name + " (" + dto.Level + ")"
}

func FormatSkillCreated(dto careerapp.SkillDTO) string {
	return fmt.Sprintf("🛠 Навык сохранён:\n<b>%s</b>", html.EscapeString(formatSkillLine(dto)))
}

func FormatSkillDeleted(dto careerapp.SkillDTO) string {
	return fmt.Sprintf("🗑 Навык удалён:\n<b>%s</b>", html.EscapeString(formatSkillLine(dto)))
}

func FormatSkillNotFound(hint string) string {
	if hint == "" {
		return "Не нашёл навык для удаления."
	}
	return fmt.Sprintf("Не нашёл навык «%s».", html.EscapeString(hint))
}

func FormatSkills(items []careerapp.SkillDTO) string {
	if len(items) == 0 {
		return "🛠 Навыков пока нет."
	}
	var b strings.Builder
	b.WriteString("🛠 <b>Навыки</b>\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, html.EscapeString(formatSkillLine(item)))
	}
	return strings.TrimSpace(b.String())
}

func FormatSkillSearchResults(query string, items []careerapp.SkillDTO) string {
	if len(items) == 0 {
		return fmt.Sprintf("По запросу «%s» навыков не найдено.", html.EscapeString(query))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "🔎 <b>Навыки: %s</b>\n", html.EscapeString(query))
	for i, item := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, html.EscapeString(formatSkillLine(item)))
	}
	return strings.TrimSpace(b.String())
}

func formatNoteTags(tags []string) string {
	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		parts = append(parts, "#"+tag)
	}
	return strings.Join(parts, " ")
}

func FormatNoteSearchResults(query string, items []knowledgeapp.NoteDTO) string {
	if len(items) == 0 {
		return fmt.Sprintf("🔍 По запросу «%s» заметок не нашёл.", html.EscapeString(query))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "🔍 <b>Заметки по запросу «%s»</b>\n", html.EscapeString(query))
	for i, item := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, html.EscapeString(truncateNote(item.Body, 80)))
	}
	return strings.TrimSpace(b.String())
}

func truncateNote(body string, max int) string {
	body = strings.TrimSpace(body)
	if len(body) <= max {
		return body
	}
	return body[:max] + "…"
}

func FormatWeightRecorded(dto healthapp.WeightLogDTO) string {
	return fmt.Sprintf("⚖️ Вес записан: <b>%.1f кг</b>", dto.WeightKg)
}

func FormatLatestWeight(dto healthapp.WeightLogDTO) string {
	return fmt.Sprintf("⚖️ Последний вес: <b>%.1f кг</b>", dto.WeightKg)
}

func FormatStepsRecorded(dto healthapp.StepLogDTO) string {
	return fmt.Sprintf("👟 Шаги записаны: <b>%d</b>", dto.Steps)
}

func FormatLatestSteps(dto healthapp.StepLogDTO) string {
	return fmt.Sprintf("👟 Последние шаги: <b>%d</b>", dto.Steps)
}

func FormatSleepRecorded(dto healthapp.SleepLogDTO) string {
	return fmt.Sprintf("😴 Сон записан: <b>%s</b>", formatSleepDuration(dto.DurationMinutes))
}

func FormatLatestSleep(dto healthapp.SleepLogDTO) string {
	return fmt.Sprintf("😴 Последний сон: <b>%s</b>", formatSleepDuration(dto.DurationMinutes))
}

func formatSleepDuration(minutes int32) string {
	h := minutes / 60
	m := minutes % 60
	if m == 0 {
		return fmt.Sprintf("%d ч", h)
	}
	return fmt.Sprintf("%d ч %d мин", h, m)
}
