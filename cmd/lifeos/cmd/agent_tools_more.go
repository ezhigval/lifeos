package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai/agent"
	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	settingsdomain "github.com/valentinezhov/lifeos/internal/settings/domain"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
)

func registerExtraAgentTools(reg *agent.Registry, d toolDeps) {
	reg.Register(agent.ToolTaskCancel,
		"Отменить задачу по названию",
		`{"title":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.cancelTitle == nil {
				return "", fmt.Errorf("cancel task unavailable")
			}
			title := argString(args, "title", "name")
			if title == "" {
				return "", fmt.Errorf("нужен title")
			}
			dto, err := d.cancelTitle.Execute(ctx, userID, title, events.SourceTelegram)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("cancelled %q", dto.Title), nil
		})

	reg.Register(agent.ToolTaskReschedule,
		"Перенести задачу на завтра по названию",
		`{"title":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.rescheduleTitle == nil {
				return "", fmt.Errorf("reschedule task unavailable")
			}
			title := argString(args, "title", "name")
			if title == "" {
				return "", fmt.Errorf("нужен title")
			}
			tomorrow, err := userTomorrow(ctx, d, userID)
			if err != nil {
				return "", err
			}
			dto, err := d.rescheduleTitle.Execute(ctx, userID, title, tomorrow, events.SourceTelegram)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("rescheduled %q to tomorrow", dto.Title), nil
		})

	reg.Register(agent.ToolTaskRescheduleAll,
		"Перенести все сегодняшние задачи на завтра",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.rescheduleAll == nil {
				return "", fmt.Errorf("reschedule all unavailable")
			}
			n, err := d.rescheduleAll.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("rescheduled %d tasks to tomorrow", n), nil
		})

	reg.Register(agent.ToolPlanSetAvailability,
		"Установить, до которого часа сегодня работаю",
		`{"hour":0-23,"minute":0-59 optional}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.setAvail == nil {
				return "", fmt.Errorf("set availability unavailable")
			}
			hour, err := argInt(args, "hour")
			if err != nil {
				return "", fmt.Errorf("нужен hour (0-23)")
			}
			minute, _ := argInt(args, "minute")
			until, err := d.setAvail.Execute(ctx, userID, hour, minute)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("available until %s", until), nil
		})

	reg.Register(agent.ToolPlanTriage,
		"Разбор перегруженного дня (что срочно / что можно перенести)",
		`{"defer_low":true optional}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.triage == nil {
				return "", fmt.Errorf("triage unavailable")
			}
			text, low, err := d.triage.Propose(ctx, userID)
			if err != nil {
				return "", err
			}
			deferLow := false
			if v, ok := args["defer_low"]; ok {
				switch t := v.(type) {
				case bool:
					deferLow = t
				case string:
					deferLow = strings.EqualFold(t, "true") || t == "1"
				}
			}
			if deferLow && len(low) > 0 {
				n, err := d.triage.ApplyDefer(ctx, userID, low)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("%s\n\ndeferred %d low-priority tasks", text, n), nil
			}
			if len(low) > 0 {
				return text + "\n\n(call again with defer_low=true to move low-priority to tomorrow)", nil
			}
			return text, nil
		})

	reg.Register(agent.ToolFinanceListPlan,
		"Показать финансовый план (плановые доходы/расходы + взносы по долгам)",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.listFinancePlan == nil {
				return "", fmt.Errorf("list finance plan unavailable")
			}
			plan, err := d.listFinancePlan.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(plan.Items) == 0 {
				return "plan is empty", nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "planned income=%d expense=%d kop\n", plan.PlannedIncome, plan.PlannedExpense)
			for _, item := range plan.Items {
				fmt.Fprintf(&b, "- [%s] %s %d kop %s next=%s\n", item.Kind, item.Title, item.AmountCents, item.Interval, item.NextDate)
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolFinanceCreatePlan,
		"Добавить плановый доход или расход",
		`{"kind":"income|expense","title":"string","amount_rubles":number,"interval":"once|weekly|monthly optional","next_date":"YYYY-MM-DD optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.createPlanned == nil {
				return "", fmt.Errorf("create planned unavailable")
			}
			kind := strings.ToLower(argString(args, "kind", "unit"))
			switch kind {
			case "доход":
				kind = "income"
			case "расход":
				kind = "expense"
			}
			if kind != "income" && kind != "expense" {
				return "", fmt.Errorf("kind must be income or expense")
			}
			title := argString(args, "title", "name")
			if title == "" {
				return "", fmt.Errorf("нужен title")
			}
			cents, err := argRublesCents(args)
			if err != nil {
				return "", err
			}
			interval := strings.ToLower(argString(args, "interval"))
			if interval == "" {
				interval = "monthly"
			}
			next, err := parsePlanNextDate(ctx, d, userID, argString(args, "next_date", "date"))
			if err != nil {
				return "", err
			}
			dto, err := d.createPlanned.Execute(ctx, financeapp.CreatePlannedCashflowInput{
				UserID: userID, Kind: kind, Title: title, AmountCents: cents,
				Interval: interval, NextDate: next, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("planned %s %q %d kop %s next=%s", dto.Kind, dto.Title, dto.AmountCents, dto.Interval, dto.NextDate), nil
		})

	reg.Register(agent.ToolHealthRecordSteps,
		"Записать шаги",
		`{"steps":number}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.recordSteps == nil {
				return "", fmt.Errorf("record steps unavailable")
			}
			steps, err := argInt(args, "steps", "count")
			if err != nil || steps <= 0 {
				return "", fmt.Errorf("нужны steps > 0")
			}
			dto, err := d.recordSteps.Execute(ctx, healthapp.RecordStepsInput{
				UserID: userID, Steps: int32(steps), Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("steps recorded %d", dto.Steps), nil
		})

	reg.Register(agent.ToolHealthLatestSteps,
		"Последние записанные шаги",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.latestSteps == nil {
				return "", fmt.Errorf("latest steps unavailable")
			}
			dto, err := d.latestSteps.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("latest steps %d", dto.Steps), nil
		})

	reg.Register(agent.ToolHealthRecordSleep,
		"Записать сон (часы или минуты)",
		`{"hours":number optional,"minutes":number optional}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.recordSleep == nil {
				return "", fmt.Errorf("record sleep unavailable")
			}
			mins, err := sleepMinutesFromArgs(args)
			if err != nil {
				return "", err
			}
			dto, err := d.recordSleep.Execute(ctx, healthapp.RecordSleepInput{
				UserID: userID, DurationMinutes: mins, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("sleep recorded %.1f h", dto.DurationHours()), nil
		})

	reg.Register(agent.ToolHealthLatestSleep,
		"Последняя запись сна",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.latestSleep == nil {
				return "", fmt.Errorf("latest sleep unavailable")
			}
			dto, err := d.latestSleep.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("latest sleep %.1f h", dto.DurationHours()), nil
		})

	reg.Register(agent.ToolCareerSkillCreate,
		"Добавить навык",
		`{"name":"string","level":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.createSkill == nil {
				return "", fmt.Errorf("create skill unavailable")
			}
			name := argString(args, "name", "title")
			if name == "" {
				return "", fmt.Errorf("нужен name")
			}
			dto, err := d.createSkill.Execute(ctx, careerapp.CreateSkillInput{
				UserID: userID, Name: name, Level: argString(args, "level"), Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("skill created %q level=%s", dto.Name, dto.Level), nil
		})

	reg.Register(agent.ToolCareerSkillList,
		"Список навыков",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.listSkills == nil {
				return "", fmt.Errorf("list skills unavailable")
			}
			items, err := d.listSkills.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no skills", nil
			}
			var b strings.Builder
			for _, s := range items {
				fmt.Fprintf(&b, "- %s", s.Name)
				if s.Level != "" {
					fmt.Fprintf(&b, " (%s)", s.Level)
				}
				b.WriteByte('\n')
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolSphereList,
		"Список сфер жизни",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.listSpheres == nil {
				return "", fmt.Errorf("list spheres unavailable")
			}
			items, err := d.listSpheres.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no spheres", nil
			}
			var b strings.Builder
			for _, s := range items {
				fmt.Fprintf(&b, "- %s\n", s.Name)
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolSphereCreate,
		"Создать сферу жизни",
		`{"name":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.createSphere == nil {
				return "", fmt.Errorf("create sphere unavailable")
			}
			name := argString(args, "name", "title")
			if name == "" {
				return "", fmt.Errorf("нужен name")
			}
			dto, err := d.createSphere.Execute(ctx, spheresapp.CreateSphereInput{
				UserID: userID, Name: name, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("sphere created %q", dto.Name), nil
		})

	reg.Register(agent.ToolQueryPriorities,
		"Что сейчас важнее всего",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.priorities == nil {
				return "", fmt.Errorf("priorities unavailable")
			}
			items, err := d.priorities.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return "no priorities", nil
			}
			var b strings.Builder
			for _, item := range items {
				fmt.Fprintf(&b, "- %s\n", item.Title)
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolAnalyticsSummary,
		"Сводка продуктивности за период",
		`{}`,
		func(ctx context.Context, userID ids.UserID, _ map[string]any) (string, error) {
			if d.analytics == nil {
				return "", fmt.Errorf("analytics unavailable")
			}
			s, err := d.analytics.Execute(ctx, userID)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s: created=%d completed=%d rate=%d%% open=%d habits=%d/%d",
				s.PeriodLabel, s.TasksCreated, s.TasksCompleted, s.CompletionRate, s.OpenTasks,
				s.HabitCompletions, s.HabitCount), nil
		})

	reg.Register(agent.ToolNoteDelete,
		"Удалить заметку по id или тексту",
		`{"id":"uuid optional","query":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.deleteNote == nil {
				return "", fmt.Errorf("delete note unavailable")
			}
			noteID, err := resolveNoteID(ctx, d, userID, argString(args, "id", "query", "title", "text"))
			if err != nil {
				return "", err
			}
			dto, err := d.deleteNote.Execute(ctx, knowledgeapp.DeleteNoteInput{
				UserID: userID, NoteID: noteID, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("note deleted id=%s", dto.ID.String()), nil
		})

	reg.Register(agent.ToolProjectArchive,
		"Архивировать проект по названию",
		`{"name":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.archiveProject == nil {
				return "", fmt.Errorf("archive project unavailable")
			}
			name := argString(args, "name", "title")
			if name == "" {
				return "", fmt.Errorf("нужен name")
			}
			dto, err := d.archiveProject.Execute(ctx, projectsapp.ArchiveProjectInput{
				UserID: userID, Name: name, Source: events.SourceTelegram,
			})
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("project archived %q", dto.Name), nil
		})

	reg.Register(agent.ToolProjectTasks,
		"Задачи проекта по названию",
		`{"name":"string"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.findProject == nil || d.listProjectTasks == nil {
				return "", fmt.Errorf("project tasks unavailable")
			}
			name := argString(args, "name", "title", "project")
			if name == "" {
				return "", fmt.Errorf("нужен name проекта")
			}
			proj, err := d.findProject.Execute(ctx, userID, name)
			if err != nil {
				return "", err
			}
			items, err := d.listProjectTasks.Execute(ctx, userID, proj.ID)
			if err != nil {
				return "", err
			}
			if len(items) == 0 {
				return fmt.Sprintf("no tasks in project %q", proj.Name), nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "project %q:\n", proj.Name)
			for i, t := range items {
				if i >= 15 {
					break
				}
				fmt.Fprintf(&b, "- %s [%s]\n", t.Title, t.Status)
			}
			return b.String(), nil
		})

	reg.Register(agent.ToolProjectProgress,
		"Прогресс проекта (по названию или первый активный)",
		`{"name":"string optional"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.projectProg == nil {
				return "", fmt.Errorf("project progress unavailable")
			}
			name := argString(args, "name", "title", "project")
			var p projectsapp.ProgressDTO
			var err error
			if name != "" {
				p, err = d.projectProg.ExecuteByName(ctx, userID, name)
			} else {
				p, err = d.projectProg.Execute(ctx, userID, ids.ProjectID{})
			}
			if err != nil {
				return "", err
			}
			if p.HasTarget {
				return fmt.Sprintf("project %q progress %s/%s (%s%%) %s", p.Name, p.Current, p.Target, p.Percent, p.Unit), nil
			}
			return fmt.Sprintf("project %q current=%s (no target)", p.Name, p.Current), nil
		})

	reg.Register(agent.ToolSettingsMorning,
		"Время утреннего обзора",
		`{"hour":0-23,"minute":0-59 optional}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.updateMorning == nil {
				return "", fmt.Errorf("morning settings unavailable")
			}
			tod, err := timeOfDayFromArgs(args)
			if err != nil {
				return "", err
			}
			at, err := d.updateMorning.Execute(ctx, userID, tod)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("morning review set to %02d:%02d", at.Hour, at.Minute), nil
		})

	reg.Register(agent.ToolSettingsEvening,
		"Время вечернего обзора",
		`{"hour":0-23,"minute":0-59 optional}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.updateEvening == nil {
				return "", fmt.Errorf("evening settings unavailable")
			}
			tod, err := timeOfDayFromArgs(args)
			if err != nil {
				return "", err
			}
			at, err := d.updateEvening.Execute(ctx, userID, tod)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("evening review set to %02d:%02d", at.Hour, at.Minute), nil
		})

	reg.Register(agent.ToolSettingsQuietHours,
		"Тихие часы (не беспокоить)",
		`{"start_hour":0-23,"start_minute":0-59 optional,"end_hour":0-23,"end_minute":0-59 optional}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			if d.updateQuiet == nil {
				return "", fmt.Errorf("quiet hours unavailable")
			}
			startH, err := argInt(args, "start_hour", "hour")
			if err != nil {
				return "", fmt.Errorf("нужен start_hour")
			}
			startM, _ := argInt(args, "start_minute", "minute")
			endH, err := argInt(args, "end_hour")
			if err != nil {
				return "", fmt.Errorf("нужен end_hour")
			}
			endM, _ := argInt(args, "end_minute")
			start := settingsdomain.TimeOfDay{Hour: startH, Minute: startM}
			end := settingsdomain.TimeOfDay{Hour: endH, Minute: endM}
			if err := d.updateQuiet.Execute(ctx, userID, start, end); err != nil {
				return "", err
			}
			return fmt.Sprintf("quiet hours %02d:%02d–%02d:%02d", startH, startM, endH, endM), nil
		})
}

func resolveNoteID(ctx context.Context, d toolDeps, userID ids.UserID, hint string) (ids.NoteID, error) {
	hint = strings.TrimSpace(hint)
	if hint != "" {
		if id, err := ids.ParseNoteID(hint); err == nil && !id.IsZero() {
			return id, nil
		}
	}
	if d.searchNotes != nil && hint != "" {
		items, err := d.searchNotes.Execute(ctx, knowledgeapp.SearchNotesInput{UserID: userID, Query: hint})
		if err != nil {
			return ids.NoteID{}, err
		}
		if len(items) > 0 {
			return items[0].ID, nil
		}
		return ids.NoteID{}, fmt.Errorf("заметка не найдена по %q", hint)
	}
	if d.listNotes == nil {
		return ids.NoteID{}, fmt.Errorf("нужен id или query")
	}
	items, err := d.listNotes.Execute(ctx, knowledgeapp.ListNotesInput{UserID: userID})
	if err != nil {
		return ids.NoteID{}, err
	}
	if len(items) == 0 {
		return ids.NoteID{}, fmt.Errorf("нет заметок")
	}
	return items[0].ID, nil
}

func timeOfDayFromArgs(args map[string]any) (settingsdomain.TimeOfDay, error) {
	hour, err := argInt(args, "hour")
	if err != nil {
		return settingsdomain.TimeOfDay{}, fmt.Errorf("нужен hour (0-23)")
	}
	minute, _ := argInt(args, "minute")
	tod := settingsdomain.TimeOfDay{Hour: hour, Minute: minute}
	if !tod.Valid() {
		return settingsdomain.TimeOfDay{}, fmt.Errorf("некорректное время")
	}
	return tod, nil
}

func userTomorrow(ctx context.Context, d toolDeps, userID ids.UserID) (time.Time, error) {
	tz := "UTC"
	if d.tzReader != nil {
		if t, err := d.tzReader.Timezone(ctx, userID); err == nil && t != "" {
			tz = t
		}
	}
	today, err := timeutil.DateInTimezone(time.Now().UTC(), tz)
	if err != nil {
		return time.Time{}, err
	}
	return today.Add(24 * time.Hour), nil
}

func parsePlanNextDate(ctx context.Context, d toolDeps, userID ids.UserID, raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		t, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return time.Time{}, fmt.Errorf("next_date must be YYYY-MM-DD")
		}
		return t.UTC(), nil
	}
	tz := "UTC"
	if d.tzReader != nil {
		if t, err := d.tzReader.Timezone(ctx, userID); err == nil && t != "" {
			tz = t
		}
	}
	return timeutil.DateInTimezone(time.Now().UTC(), tz)
}

func sleepMinutesFromArgs(args map[string]any) (int32, error) {
	if mins, err := argInt(args, "minutes", "duration_minutes"); err == nil && mins > 0 {
		return int32(mins), nil
	}
	if h, err := argFloat(args, "hours", "duration_hours", "h"); err == nil && h > 0 {
		return int32(h * 60), nil
	}
	return 0, fmt.Errorf("нужны hours или minutes")
}

func argInt(args map[string]any, keys ...string) (int, error) {
	for _, k := range keys {
		if v, ok := args[k]; ok && v != nil {
			switch t := v.(type) {
			case float64:
				return int(t), nil
			case int:
				return t, nil
			case int64:
				return int(t), nil
			case string:
				n, err := parseIntLoose(t)
				if err == nil {
					return n, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("нужно целое число")
}

func parseIntLoose(s string) (int, error) {
	s = strings.TrimSpace(s)
	n := 0
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not int")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
