package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai/agent"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	memoryapp "github.com/valentinezhov/lifeos/internal/memory/app"
	memorydomain "github.com/valentinezhov/lifeos/internal/memory/domain"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

// toolDeps is the subset of runtime use cases exposed to the conversational agent.
type toolDeps struct {
	createTask    *tasksapp.CreateTask
	listToday     *tasksapp.ListTasksToday
	completeTitle *tasksapp.CompleteTaskByTitle
	recordExpense *financeapp.RecordExpense
	recordIncome  *financeapp.RecordIncome
	reminder      *notifapp.ScheduleReminder
	createHabit   *habitsapp.CreateHabit
	trackHabit    *habitsapp.TrackHabit
	upsertMemory  *memoryapp.UpsertMemory
	recallMemory  *memoryapp.Recall
	tzReader      interface {
		Timezone(ctx context.Context, userID ids.UserID) (string, error)
	}
}

func registerAgentTools(reg *agent.Registry, d toolDeps) {
	reg.Register(agent.ToolTaskCreate,
		"Создать задачу на сегодня",
		`{"title":"string","priority":"low|medium|high optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			title := strings.TrimSpace(argString(args, "title", "name"))
			if title == "" {
				return "", fmt.Errorf("нужен title — конкретная формулировка из запроса пользователя")
			}
			low := strings.ToLower(title)
			if low == "разобрать входящие" || low == "разобрать почту" || low == "задача" {
				return "", fmt.Errorf("title выглядит как заглушка (%q); переформулируй из слов пользователя или спроси уточнение", title)
			}
			prio := taskdomain.PriorityMedium
			if p := argString(args, "priority"); p != "" {
				prio = taskdomain.Priority(p)
			}
			today := time.Now().UTC().Truncate(24 * time.Hour)
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

	reg.Register(agent.ToolReminderCreate,
		"Создать напоминание",
		`{"message":"string","time_text":"вечером|через час|в 18:30"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			msg := argString(args, "message", "title", "text")
			if msg == "" {
				return "", fmt.Errorf("нужен message")
			}
			tz := "UTC"
			if d.tzReader != nil {
				if t, err := d.tzReader.Timezone(ctx, userID); err == nil && t != "" {
					tz = t
				}
			}
			fireAt := rulebased.ParseFireAt(time.Now().UTC(), tz, argString(args, "time_text", "when"))
			if fireAt.Before(time.Now().UTC()) {
				fireAt = time.Now().UTC().Add(time.Hour)
			}
			dto, err := d.reminder.Execute(ctx, notifapp.ScheduleReminderInput{
				UserID: userID, Message: msg, FireAt: fireAt,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("reminder scheduled %q at %s", dto.Message, dto.FireAt.Format(time.RFC3339)), nil
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
