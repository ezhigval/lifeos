package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/agent"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
	calendarapp "github.com/valentinezhov/lifeos/internal/calendar/app"
	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	memoryapp "github.com/valentinezhov/lifeos/internal/memory/app"
	memorydomain "github.com/valentinezhov/lifeos/internal/memory/domain"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	planapp "github.com/valentinezhov/lifeos/internal/planning/app"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	"github.com/valentinezhov/lifeos/internal/query"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	spheresdomain "github.com/valentinezhov/lifeos/internal/spheres/domain"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

// toolDeps is the subset of runtime use cases exposed to the conversational agent.
type toolDeps struct {
	createTask       *tasksapp.CreateTask
	listToday        *tasksapp.ListTasksToday
	completeTitle    *tasksapp.CompleteTaskByTitle
	cancelTitle      *tasksapp.CancelTaskByTitle
	rescheduleTitle  *tasksapp.RescheduleTaskByTitle
	rescheduleAll    *planapp.RescheduleTasks
	setAvail         *planapp.SetDayAvailability
	triage           *planapp.TriageOverloadedDay
	recordExpense    *financeapp.RecordExpense
	recordIncome     *financeapp.RecordIncome
	listDebts        *financeapp.ListDebts
	createDebt       *financeapp.CreateDebt
	payDebt          *financeapp.PayDebt
	cashFlow         *financeapp.CashFlowSummary
	listFinancePlan  *financeapp.ListFinancePlan
	createPlanned    *financeapp.CreatePlannedCashflow
	reminder         *notifapp.ScheduleReminder
	listReminders    *notifapp.ListReminders
	cancelReminder   *notifapp.CancelReminder
	createHabit      *habitsapp.CreateHabit
	trackHabit       *habitsapp.TrackHabit
	listHabits       *habitsapp.ListHabitsToday
	createNote       *knowledgeapp.CreateNote
	listNotes        *knowledgeapp.ListNotes
	searchNotes      *knowledgeapp.SearchNotes
	createEvent      *calendarapp.CreateEvent
	listCalendar     *calendarapp.ListEventsToday
	createProject    *projectsapp.CreateProject
	listProjects     *projectsapp.ListProjects
	listSpheres      *spheresapp.ListSpheres
	findSphere       *spheresapp.FindSphereByName
	createSphere     *spheresapp.CreateSphere
	recordWeight     *healthapp.RecordWeight
	latestWeight     *healthapp.GetLatestWeight
	recordSteps      *healthapp.RecordSteps
	latestSteps      *healthapp.GetLatestSteps
	recordSleep      *healthapp.RecordSleep
	latestSleep      *healthapp.GetLatestSleep
	createContact    *careerapp.CreateContact
	listContacts     *careerapp.ListContacts
	createSkill      *careerapp.CreateSkill
	listSkills       *careerapp.ListSkills
	priorities       *query.GetTopPriorities
	analytics        *query.GetProductivitySummary
	upsertMemory     *memoryapp.UpsertMemory
	recallMemory     *memoryapp.Recall
	tzReader         interface {
		Timezone(ctx context.Context, userID ids.UserID) (string, error)
	}
}

func registerAgentTools(reg *agent.Registry, d toolDeps) {
	reg.Register(agent.ToolTaskCreate,
		"Создать задачу на сегодня",
		`{"title":"string","priority":"low|medium|high optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			title := strings.TrimSpace(argString(args, "title", "name"))
			if ai.IsPlaceholderTitle(title) {
				return "", fmt.Errorf("title выглядит как заглушка (%q); переформулируй из слов пользователя или спроси уточнение", title)
			}
			prio := taskdomain.PriorityMedium
			if p := argString(args, "priority"); p != "" {
				prio = taskdomain.Priority(p)
			}
			tz := "UTC"
			if d.tzReader != nil {
				if t, err := d.tzReader.Timezone(ctx, userID); err == nil && t != "" {
					tz = t
				}
			}
			today, err := timeutil.DateInTimezone(time.Now().UTC(), tz)
			if err != nil {
				return "", err
			}
			dto, err := d.createTask.Execute(ctx, tasksapp.CreateTaskInput{
				UserID: userID, Title: title, Priority: prio, DueDate: &today, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("created task id=%s title=%q", dto.ID.String(), dto.Title), nil
		})

	reg.Register(agent.ToolTaskListToday,
		"Список задач на сегодня",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			items, err := d.listToday.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no tasks today", nil
			}
			var b strings.Builder
			for i, t := range items {
				if i >= 15 {
					fmt.Fprintf(&b, "…and %d more", len(items)-15)
					break
				}
				fmt.Fprintf(&b, "- %s [%s]\n", t.Title, t.Status)
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolTaskComplete,
		"Отметить задачу выполненной по названию",
		`{"title":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			title := argString(args, "title", "name")
			if title == "" {
				return "", fmt.Errorf("нужен title")
			}
			dto, err := d.completeTitle.Execute(ctx, userID, title, events.SourceTelegram)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("completed %q", dto.Title), nil
		})

	reg.Register(agent.ToolFinanceExpense,
		"Записать расход",
		`{"amount_rubles":number,"category":"string optional","description":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			cents, err := argRublesCents(args)
			if err != nil {
				return "", err
			}
			dto, err := d.recordExpense.Execute(ctx, financeapp.RecordExpenseInput{
				UserID: userID, AmountCents: cents, Currency: "RUB",
				CategoryName: argString(args, "category", "title"),
				Description:  argString(args, "description"),
				Source:       events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("expense recorded %d kop category=%s", dto.AmountCents, dto.Description), nil
		})

	reg.Register(agent.ToolFinanceIncome,
		"Записать доход",
		`{"amount_rubles":number,"description":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			cents, err := argRublesCents(args)
			if err != nil {
				return "", err
			}
			dto, err := d.recordIncome.Execute(ctx, financeapp.RecordIncomeInput{
				UserID: userID, AmountCents: cents, Currency: "RUB",
				Description: argString(args, "description", "title"),
				Source:      events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("income recorded %d kop", dto.AmountCents), nil
		})

	reg.Register(agent.ToolFinanceListDebts,
		"Список открытых долгов",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.listDebts == nil {
				return "", fmt.Errorf("list debts unavailable")
			}
			items, err := d.listDebts.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no open debts", nil
			}
			var b strings.Builder
			for _, debt := range items {
				fmt.Fprintf(&b, "- %s remaining=%d kop\n", debt.Creditor, debt.RemainingCents)
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolFinanceCreateDebt,
		"Создать долг",
		`{"creditor":"string","amount_rubles":number}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.createDebt == nil {
				return "", fmt.Errorf("create debt unavailable")
			}
			creditor := argString(args, "creditor", "target", "name")
			if creditor == "" {
				return "", fmt.Errorf("нужен creditor")
			}
			cents, err := argRublesCents(args)
			if err != nil {
				return "", err
			}
			dto, err := d.createDebt.Execute(ctx, financeapp.CreateDebtInput{
				UserID: userID, Creditor: creditor, AmountCents: cents, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("debt created creditor=%q amount=%d kop", dto.Creditor, dto.AmountCents), nil
		})

	reg.Register(agent.ToolFinancePayDebt,
		"Погасить долг (частично или полностью) по кредитору",
		`{"creditor":"string","amount_rubles":number}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.payDebt == nil {
				return "", fmt.Errorf("pay debt unavailable")
			}
			creditor := argString(args, "creditor", "target", "name")
			if creditor == "" {
				return "", fmt.Errorf("нужен creditor")
			}
			cents, err := argRublesCents(args)
			if err != nil {
				return "", err
			}
			dto, err := d.payDebt.Execute(ctx, financeapp.PayDebtInput{
				UserID: userID, Creditor: creditor, AmountCents: cents, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("debt paid creditor=%q remaining=%d kop", dto.Creditor, dto.RemainingCents), nil
		})

	reg.Register(agent.ToolFinanceCashFlow,
		"Сводка доходов/расходов за текущий месяц",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.cashFlow == nil {
				return "", fmt.Errorf("cash flow unavailable")
			}
			dto, err := d.cashFlow.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("month income=%d expense=%d net=%d kop", dto.IncomeCents, dto.ExpenseCents, dto.NetCents), nil
		})

	reg.Register(agent.ToolReminderCreate,
		"Создать напоминание",
		`{"message":"string","time_text":"вечером|через час|в 18:30"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			msg := argString(args, "message", "title", "text")
			if msg == "" {
				return "", fmt.Errorf("нужен message")
			}
			timeText := argString(args, "time_text", "when")
			if timeText == "" {
				return "", fmt.Errorf("нужен time_text (когда напомнить); иначе спроси уточнение")
			}
			tz := "UTC"
			if d.tzReader != nil {
				if t, err := d.tzReader.Timezone(ctx, userID); err == nil && t != "" {
					tz = t
				}
			}
			now := time.Now().UTC()
			fireAt := rulebased.EnsureFutureFireAt(rulebased.ParseFireAt(now, tz, timeText), now)
			dto, err := d.reminder.Execute(ctx, notifapp.ScheduleReminderInput{
				UserID: userID, Message: msg, FireAt: fireAt,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("reminder scheduled %q at %s", dto.Message, dto.FireAt.Format(time.RFC3339)), nil
		})

	reg.Register(agent.ToolReminderCancel,
		"Отменить напоминание по тексту сообщения",
		`{"message":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.listReminders == nil || d.cancelReminder == nil {
				return "", fmt.Errorf("cancel reminder unavailable")
			}
			hint := strings.ToLower(argString(args, "message", "title", "text"))
			items, err := d.listReminders.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "", fmt.Errorf("нет активных напоминаний")
			}
			var pick notifapp.ReminderDTO
			if hint == "" {
				pick = items[0]
			} else {
				found := false
				for _, item := range items {
					if strings.Contains(strings.ToLower(item.Message), hint) {
						pick = item
						found = true
						break
					}
				}
				if !found {
					return "", fmt.Errorf("напоминание не найдено по %q", hint)
				}
			}
			id, err := uuid.Parse(pick.ID)
			if err != nil {
				return "", err
			}
			dto, err := d.cancelReminder.Execute(ctx, notifapp.CancelReminderInput{
				UserID: userID, ReminderID: id,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("reminder cancelled %q", dto.Message), nil
		})

	reg.Register(agent.ToolHabitCreate,
		"Создать привычку",
		`{"name":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			name := argString(args, "name", "title")
			if name == "" {
				return "", fmt.Errorf("нужен name")
			}
			dto, err := d.createHabit.Execute(ctx, habitsapp.CreateHabitInput{
				UserID: userID, Name: name, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("habit created %q", dto.Name), nil
		})

	reg.Register(agent.ToolHabitTrack,
		"Отметить привычку за сегодня",
		`{"name":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			name := argString(args, "name", "title")
			if name == "" {
				return "", fmt.Errorf("нужен name")
			}
			dto, err := d.trackHabit.Execute(ctx, habitsapp.TrackHabitInput{
				UserID: userID, Name: name, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("habit tracked %q", dto.Name), nil
		})

	reg.Register(agent.ToolHabitList,
		"Список привычек на сегодня",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.listHabits == nil {
				return "", fmt.Errorf("list habits unavailable")
			}
			items, err := d.listHabits.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no habits", nil
			}
			var b strings.Builder
			for _, h := range items {
				done := "open"
				if h.TodayCompleted {
					done = "done"
				}
				fmt.Fprintf(&b, "- %s [%s]\n", h.Name, done)
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolNoteCreate,
		"Создать заметку",
		`{"body":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.createNote == nil {
				return "", fmt.Errorf("create note unavailable")
			}
			body := argString(args, "body", "text", "title")
			if body == "" {
				return "", fmt.Errorf("нужен body")
			}
			dto, err := d.createNote.Execute(ctx, knowledgeapp.CreateNoteInput{
				UserID: userID, Body: body, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("note created id=%s", dto.ID.String()), nil
		})

	reg.Register(agent.ToolNoteList,
		"Список недавних заметок",
		`{"tag":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.listNotes == nil {
				return "", fmt.Errorf("list notes unavailable")
			}
			items, err := d.listNotes.Execute(ctx, knowledgeapp.ListNotesInput{
				UserID: userID, Tag: argString(args, "tag"),
			})
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no notes", nil
			}
			var b strings.Builder
			for i, n := range items {
				if i >= 10 {
					break
				}
				fmt.Fprintf(&b, "- %s\n", truncateRunes(n.Body, 80))
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolNoteSearch,
		"Поиск заметок",
		`{"query":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.searchNotes == nil {
				return "", fmt.Errorf("search notes unavailable")
			}
			q := argString(args, "query", "q", "text")
			if q == "" {
				return "", fmt.Errorf("нужен query")
			}
			items, err := d.searchNotes.Execute(ctx, knowledgeapp.SearchNotesInput{UserID: userID, Query: q})
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no notes matched", nil
			}
			var b strings.Builder
			for i, n := range items {
				if i >= 10 {
					break
				}
				fmt.Fprintf(&b, "- %s\n", truncateRunes(n.Body, 80))
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolCalendarCreate,
		"Создать событие календаря",
		`{"title":"string","time_text":"вечером|завтра|утром"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.createEvent == nil {
				return "", fmt.Errorf("create event unavailable")
			}
			title := argString(args, "title", "name")
			if ai.IsPlaceholderTitle(title) {
				return "", fmt.Errorf("нужен конкретный title события")
			}
			tz := "UTC"
			if d.tzReader != nil {
				if t, err := d.tzReader.Timezone(ctx, userID); err == nil && t != "" {
					tz = t
				}
			}
			now := time.Now().UTC()
			starts := rulebased.EnsureFutureFireAt(rulebased.ParseFireAt(now, tz, argString(args, "time_text", "when")), now)
			dto, err := d.createEvent.Execute(ctx, calendarapp.CreateEventInput{
				UserID: userID, Title: title, StartsAt: starts, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("event created %q at %s", dto.Title, dto.StartsAt.Format(time.RFC3339)), nil
		})

	reg.Register(agent.ToolCalendarListToday,
		"События на сегодня",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.listCalendar == nil {
				return "", fmt.Errorf("list calendar unavailable")
			}
			items, err := d.listCalendar.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no events today", nil
			}
			var b strings.Builder
			for _, e := range items {
				fmt.Fprintf(&b, "- %s @ %s\n", e.Title, e.StartsAt.Format(time.RFC3339))
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolProjectCreate,
		"Создать проект",
		`{"name":"string","sphere":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.createProject == nil {
				return "", fmt.Errorf("create project unavailable")
			}
			name := argString(args, "name", "title")
			if name == "" {
				return "", fmt.Errorf("нужен name")
			}
			sphereIDs, err := resolveToolSphereIDs(ctx, d, userID, argString(args, "sphere", "target"))
			if err != nil {
				return "", err
			}
			dto, err := d.createProject.Execute(ctx, projectsapp.CreateProjectInput{
				UserID: userID, Name: name, SphereIDs: sphereIDs, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("project created %q", dto.Name), nil
		})

	reg.Register(agent.ToolProjectList,
		"Список активных проектов",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.listProjects == nil {
				return "", fmt.Errorf("list projects unavailable")
			}
			items, err := d.listProjects.Execute(ctx, projectsapp.ListProjectsInput{UserID: userID})
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no projects", nil
			}
			var b strings.Builder
			for _, p := range items {
				fmt.Fprintf(&b, "- %s\n", p.Name)
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolHealthRecordWeight,
		"Записать вес (кг)",
		`{"weight_kg":number}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.recordWeight == nil {
				return "", fmt.Errorf("record weight unavailable")
			}
			kg, err := argFloat(args, "weight_kg", "weight", "kg")
			if err != nil {
				return "", err
			}
			dto, err := d.recordWeight.Execute(ctx, healthapp.RecordWeightInput{
				UserID: userID, WeightKg: kg, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("weight recorded %.1f kg", dto.WeightKg), nil
		})

	reg.Register(agent.ToolHealthLatestWeight,
		"Последний записанный вес",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.latestWeight == nil {
				return "", fmt.Errorf("latest weight unavailable")
			}
			dto, err := d.latestWeight.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("latest weight %.1f kg", dto.WeightKg), nil
		})

	reg.Register(agent.ToolCareerContactCreate,
		"Сохранить контакт (карьера)",
		`{"name":"string","company":"string optional","role":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.createContact == nil {
				return "", fmt.Errorf("create contact unavailable")
			}
			name := argString(args, "name", "title")
			if name == "" {
				return "", fmt.Errorf("нужен name")
			}
			dto, err := d.createContact.Execute(ctx, careerapp.CreateContactInput{
				UserID: userID, Name: name,
				Company: argString(args, "company"),
				Role:    argString(args, "role"),
				Notes:   argString(args, "notes"),
				Source:  events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("contact created %q", dto.Name), nil
		})

	reg.Register(agent.ToolCareerContactList,
		"Список контактов",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.listContacts == nil {
				return "", fmt.Errorf("list contacts unavailable")
			}
			items, err := d.listContacts.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no contacts", nil
			}
			var b strings.Builder
			for _, c := range items {
				fmt.Fprintf(&b, "- %s", c.Name)
				if c.Company != "" {
					fmt.Fprintf(&b, " (%s)", c.Company)
				}
				b.WriteByte('\n')
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolMemorySave,
		"Сохранить персональный факт/предпочтение пользователя (только для него)",
		`{"kind":"preference|fact|alias|pattern","key":"string","value":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			kind := memorydomain.Kind(argString(args, "kind"))
			if kind == "" {
				kind = memorydomain.KindFact
			}
			key := argString(args, "key")
			val := argString(args, "value")
			mem, err := d.upsertMemory.Execute(ctx, memoryapp.UpsertMemoryInput{
				UserID: userID, Kind: kind, Key: key, Value: val, Source: "agent",
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("memory saved kind=%s key=%s", mem.Kind, mem.Key), nil
		})

	reg.Register(agent.ToolMemoryRecall,
		"Вспомнить сохранённые факты о пользователе",
		`{"query":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			items, err := d.recallMemory.Execute(ctx, userID, argString(args, "query"), 10)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no memories", nil
			}
			var b strings.Builder
			for _, m := range items {
				fmt.Fprintf(&b, "- [%s] %s = %s\n", m.Kind, m.Key, m.Value)
			}
			return b.String(), nil
		})

	registerExtraAgentTools(reg, d)
}

func resolveToolSphereIDs(ctx context.Context, d toolDeps, userID ids.UserID, sphereName string) ([]ids.SphereID, error) {
	sphereName = strings.TrimSpace(sphereName)
	if sphereName != "" && d.findSphere != nil {
		s, err := d.findSphere.Execute(ctx, userID, sphereName)
		if err != nil {
			return nil, err
		}
		return []ids.SphereID{s.ID}, nil
	}
	if d.listSpheres != nil {
		items, err := d.listSpheres.Execute(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, want := range spheresdomain.DefaultSphereNames {
			for _, s := range items {
				if strings.EqualFold(s.Name, want) {
					return []ids.SphereID{s.ID}, nil
				}
			}
		}
		if len(items) > 0 {
			return []ids.SphereID{items[0].ID}, nil
		}
	}
	return nil, fmt.Errorf("нет сфер — создай сферу сначала")
}

func argString(args map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := args[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				if s := strings.TrimSpace(t); s != "" {
					return s
				}
			case float64:
				return strconv.FormatFloat(t, 'f', -1, 64)
			case int:
				return strconv.Itoa(t)
			case int64:
				return strconv.FormatInt(t, 10)
			}
		}
	}
	return ""
}

func argFloat(args map[string]any, keys ...string) (float64, error) {
	for _, k := range keys {
		if v, ok := args[k]; ok && v != nil {
			switch t := v.(type) {
			case float64:
				return t, nil
			case int:
				return float64(t), nil
			case int64:
				return float64(t), nil
			case string:
				f, err := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(t, ",", ".")), 64)
				if err == nil {
					return f, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("нужно числовое значение")
}

func argRublesCents(args map[string]any) (int64, error) {
	if v, ok := args["amount_cents"]; ok {
		switch t := v.(type) {
		case float64:
			return int64(t), nil
		case int64:
			return t, nil
		case int:
			return int64(t), nil
		case string:
			n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
			if err == nil {
				return n, nil
			}
		}
	}
	if v, ok := args["amount_rubles"]; ok {
		switch t := v.(type) {
		case float64:
			return int64(t * 100), nil
		case int64:
			return t * 100, nil
		case int:
			return int64(t) * 100, nil
		case string:
			if cents, err := rulebased.ParseRublesAmount(t); err == nil {
				return cents, nil
			}
		}
	}
	if s := argString(args, "amount", "sum"); s != "" {
		if cents, err := rulebased.ParseRublesAmount(s); err == nil {
			return cents, nil
		}
	}
	return 0, fmt.Errorf("нужна сумма amount_rubles")
}

func truncateRunes(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max]) + "…"
}
