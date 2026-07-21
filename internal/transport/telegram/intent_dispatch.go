package telegram

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
	calendarapp "github.com/valentinezhov/lifeos/internal/calendar/app"
	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	careerdomain "github.com/valentinezhov/lifeos/internal/career/domain"
	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	financedomain "github.com/valentinezhov/lifeos/internal/finance/domain"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	habitsdomain "github.com/valentinezhov/lifeos/internal/habits/domain"
	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	"github.com/valentinezhov/lifeos/internal/health/domain"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	projectsdomain "github.com/valentinezhov/lifeos/internal/projects/domain"
	settingsdomain "github.com/valentinezhov/lifeos/internal/settings/domain"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	spheresdomain "github.com/valentinezhov/lifeos/internal/spheres/domain"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

func (h *MessageHandler) dispatchIntent(ctx context.Context, userID ids.UserID, intent ai.ResolvedIntent) (dispatchResult, error) {
	switch intent.Type {
	case ai.IntentTaskCreate:
		if ai.IsPlaceholderTitle(intent.Title) {
			return dispatchResult{text: "Не понял название задачи. Например: «добавь задачу купить фильтр»."}, nil
		}
		tz, err := h.tzReader.Timezone(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		today, err := timeutil.DateInTimezone(time.Now().UTC(), tz)
		if err != nil {
			return dispatchResult{}, err
		}
		projectIDs, err := h.resolveProjectIDs(ctx, userID, intent.Unit, intent.Target)
		if err != nil {
			if errors.Is(err, projectsdomain.ErrNotFound) {
				return dispatchResult{text: FormatProjectNotFound(intent.Target)}, nil
			}
			return dispatchResult{}, err
		}
		dto, err := h.createTask.Execute(ctx, tasksapp.CreateTaskInput{
			UserID: userID, Title: intent.Title, Priority: taskdomain.PriorityMedium,
			DueDate: &today, ProjectIDs: projectIDs, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatTaskCreated(dto)}, err
	case ai.IntentTaskListToday:
		return h.tasksTodayView(ctx, userID)
	case ai.IntentTaskComplete:
		dto, err := h.completeByTitle.Execute(ctx, userID, intent.Title, events.SourceTelegram)
		if errors.Is(err, tasksapp.ErrTaskNotFound) {
			return dispatchResult{text: FormatTaskNotFound(intent.Title)}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatTaskCompleted(dto)}, nil
	case ai.IntentTaskCancel:
		dto, err := h.cancelByTitle.Execute(ctx, userID, intent.Title, events.SourceTelegram)
		if errors.Is(err, tasksapp.ErrTaskNotFound) {
			return dispatchResult{text: FormatTaskNotFound(intent.Title)}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatTaskCancelled(dto)}, nil
	case ai.IntentTaskRescheduleOne:
		tz, err := h.tzReader.Timezone(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		today, err := timeutil.DateInTimezone(time.Now().UTC(), tz)
		if err != nil {
			return dispatchResult{}, err
		}
		tomorrow := today.Add(24 * time.Hour)
		dto, err := h.rescheduleByTitle.Execute(ctx, userID, intent.Title, tomorrow, events.SourceTelegram)
		if errors.Is(err, tasksapp.ErrTaskNotFound) {
			return dispatchResult{text: FormatTaskNotFound(intent.Title)}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatTaskRescheduled(dto)}, nil
	case ai.IntentTaskListByTag:
		items, err := h.listByTag.Execute(ctx, userID, intent.Title)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatTasksByTag(intent.Title, items)}, nil
	case ai.IntentProjectProgress:
		var p projectsapp.ProgressDTO
		var err error
		if strings.TrimSpace(intent.Title) != "" {
			p, err = h.projectProg.ExecuteByName(ctx, userID, intent.Title)
		} else {
			p, err = h.projectProg.Execute(ctx, userID, ids.ProjectID{})
		}
		return dispatchResult{text: FormatProjectProgress(p)}, err
	case ai.IntentQueryPriorities:
		items, err := h.priorities.Execute(ctx, userID)
		return dispatchResult{text: FormatPriorities(items)}, err
	case ai.IntentReminderCreate:
		msg := strings.TrimSpace(intent.Message)
		if msg == "" {
			return dispatchResult{text: "Не понял текст напоминания. Например: «напомни вечером позвонить»."}, nil
		}
		tz, err := h.tzReader.Timezone(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		fireAt := rulebased.EnsureFutureFireAt(rulebased.ParseFireAt(time.Now().UTC(), tz, intent.TimeText), time.Now().UTC())
		dto, err := h.reminder.Execute(ctx, notifapp.ScheduleReminderInput{
			UserID: userID, Message: msg, FireAt: fireAt,
		})
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatReminderScheduled(dto.Message, formatLocalTime(dto.FireAt, tz))}, nil
	case ai.IntentReminderCancel:
		tz, err := h.tzReader.Timezone(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		reminderID, hint, err := h.resolveReminderToCancel(ctx, userID, intent.Message)
		if err != nil {
			if errors.Is(err, notifapp.ErrReminderNotFound) {
				return dispatchResult{text: FormatReminderNotFound(hint)}, nil
			}
			return dispatchResult{}, err
		}
		dto, err := h.cancelReminder.Execute(ctx, notifapp.CancelReminderInput{
			UserID: userID, ReminderID: reminderID,
		})
		if err != nil {
			if errors.Is(err, notifapp.ErrReminderNotFound) {
				return dispatchResult{text: FormatReminderNotFound(hint)}, nil
			}
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatReminderCancelled(dto.Message, formatLocalTime(dto.FireAt, tz))}, nil
	case ai.IntentPlanSetAvailability:
		until, err := h.setAvail.Execute(ctx, userID, intent.Hour, intent.Minute)
		return dispatchResult{text: FormatAvailability(until)}, err
	case ai.IntentPlanTriage:
		return h.runAction(ctx, userID, ActionTriage)
	case ai.IntentTaskReschedule:
		n, err := h.reschedule.Execute(ctx, userID)
		return dispatchResult{text: fmt.Sprintf("Перенесено задач на завтра: %d", n)}, err
	case ai.IntentSettingsMorning:
		at, err := h.updateMorning.Execute(ctx, userID, settingsdomain.TimeOfDay{Hour: intent.Hour, Minute: intent.Minute})
		return dispatchResult{text: FormatMorningReviewSet(at)}, err
	case ai.IntentSettingsEvening:
		at, err := h.updateEvening.Execute(ctx, userID, settingsdomain.TimeOfDay{Hour: intent.Hour, Minute: intent.Minute})
		return dispatchResult{text: FormatEveningReviewSet(at)}, err
	case ai.IntentSettingsQuietHours:
		endHour := atoiIntent(intent.Target)
		endMin := atoiIntent(intent.Unit)
		err := h.updateQuiet.Execute(ctx, userID,
			settingsdomain.TimeOfDay{Hour: intent.Hour, Minute: intent.Minute},
			settingsdomain.TimeOfDay{Hour: endHour, Minute: endMin},
		)
		return dispatchResult{text: FormatQuietHoursSet(intent.Hour, intent.Minute, endHour, endMin)}, err
	case ai.IntentFinanceIncome:
		if intent.AmountCents <= 0 {
			return dispatchResult{text: "Не понял сумму. Пример: «пришёл заказ на 50 тысяч»"}, nil
		}
		desc := intent.Title
		if desc == "" {
			desc = "доход"
		}
		dto, err := h.recordIncome.Execute(ctx, financeapp.RecordIncomeInput{
			UserID: userID, AmountCents: intent.AmountCents, Currency: intent.Currency,
			Description: desc, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatIncomeRecorded(dto)}, err
	case ai.IntentFinanceExpense:
		if intent.AmountCents <= 0 {
			return dispatchResult{text: "Не понял сумму. Пример: «потратил 5 тысяч на еду»"}, nil
		}
		dto, err := h.recordExpense.Execute(ctx, financeapp.RecordExpenseInput{
			UserID: userID, AmountCents: intent.AmountCents, Currency: intent.Currency,
			CategoryName: intent.Title, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatExpenseRecorded(dto)}, err
	case ai.IntentFinanceListDebts:
		items, err := h.listDebts.Execute(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatDebts(items)}, nil
	case ai.IntentFinanceCashFlow:
		summary, err := h.cashFlow.Execute(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatCashFlow(summary)}, nil
	case ai.IntentFinanceCreateDebt:
		if intent.AmountCents <= 0 || intent.Target == "" {
			return dispatchResult{text: "Пример: «долг 100 тысяч банку»"}, nil
		}
		dto, err := h.createDebt.Execute(ctx, financeapp.CreateDebtInput{
			UserID: userID, Creditor: intent.Target, AmountCents: intent.AmountCents,
			Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatDebtCreated(dto)}, err
	case ai.IntentFinancePayDebt:
		if intent.AmountCents <= 0 || intent.Target == "" {
			return dispatchResult{text: "Пример: «заплатил 10 тысяч банку»"}, nil
		}
		dto, err := h.payDebt.Execute(ctx, financeapp.PayDebtInput{
			UserID: userID, Creditor: intent.Target, AmountCents: intent.AmountCents,
			Source: events.SourceTelegram,
		})
		if errors.Is(err, financedomain.ErrDebtNotFound) {
			return dispatchResult{text: fmt.Sprintf("Не нашёл открытый долг «%s»", intent.Target)}, nil
		}
		if errors.Is(err, financedomain.ErrOverpayment) {
			return dispatchResult{text: "Сумма больше остатка по долгу"}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatDebtPaid(dto, intent.AmountCents)}, nil
	case ai.IntentHabitCreate:
		dto, err := h.createHabit.Execute(ctx, habitsapp.CreateHabitInput{
			UserID: userID, Name: intent.Title, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatHabitCreated(dto)}, err
	case ai.IntentHabitTrack:
		if intent.Title == "" {
			return dispatchResult{text: "Пример: «отметь привычку бег»"}, nil
		}
		result, err := h.trackHabit.Execute(ctx, habitsapp.TrackHabitInput{
			UserID: userID, Name: intent.Title, Source: events.SourceTelegram,
		})
		if errors.Is(err, habitsdomain.ErrNotFound) {
			return dispatchResult{text: FormatHabitNotFound(intent.Title)}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		today, err := h.habitsTodayView(ctx, userID)
		if err != nil {
			return dispatchResult{text: FormatHabitTracked(result)}, nil
		}
		return dispatchResult{text: FormatHabitTracked(result) + "\n\n" + today.text, inline: today.inline}, nil
	case ai.IntentHabitList:
		return h.habitsTodayView(ctx, userID)
	case ai.IntentProjectCreate:
		sphereIDs, err := h.resolveSphereIDs(ctx, userID, intent.Unit, intent.Target)
		if err != nil {
			if errors.Is(err, spheresdomain.ErrNotFound) {
				return dispatchResult{text: FormatSphereNotFound(intent.Target)}, nil
			}
			return dispatchResult{}, err
		}
		dto, err := h.createProject.Execute(ctx, projectsapp.CreateProjectInput{
			UserID: userID, Name: intent.Title, SphereIDs: sphereIDs, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatProjectCreated(dto)}, err
	case ai.IntentProjectList:
		if intent.Target != "" {
			return h.projectsBySphereView(ctx, userID, intent.Target)
		}
		return h.projectsPickerView(ctx, userID)
	case ai.IntentProjectTasks:
		if intent.Title == "" {
			return dispatchResult{text: "Пример: «задачи проекта свадьба»"}, nil
		}
		return h.projectTasksByName(ctx, userID, intent.Title)
	case ai.IntentProjectArchive:
		if intent.Title == "" {
			return dispatchResult{text: "Пример: «архивируй проект свадьба»"}, nil
		}
		dto, err := h.archiveProject.Execute(ctx, projectsapp.ArchiveProjectInput{
			UserID: userID, Name: intent.Title, Source: events.SourceTelegram,
		})
		if errors.Is(err, projectsdomain.ErrNotFound) {
			return dispatchResult{text: FormatProjectNotFound(intent.Title)}, nil
		}
		if errors.Is(err, projectsdomain.ErrNotActive) {
			return dispatchResult{text: fmt.Sprintf("Проект «%s» уже не активен", intent.Title)}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatProjectArchived(dto)}, nil
	case ai.IntentCalendarCreate:
		if intent.Title == "" {
			return dispatchResult{text: "Пример: «добавь встречу с дизайнером завтра в 15»"}, nil
		}
		tz, err := h.tzReader.Timezone(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		startsAt, err := rulebased.ParseEventStart(time.Now().UTC(), tz, intent.Target, intent.Hour, intent.Minute)
		if err != nil {
			return dispatchResult{}, err
		}
		dto, err := h.createEvent.Execute(ctx, calendarapp.CreateEventInput{
			UserID: userID, Title: intent.Title, StartsAt: startsAt, Source: events.SourceTelegram,
		})
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatCalendarEventCreated(dto, tz)}, nil
	case ai.IntentCalendarListToday:
		return h.calendarTodayView(ctx, userID)
	case ai.IntentReviewWeekly:
		text, err := h.review.Weekly(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatAssistantHTML(text)}, nil
	case ai.IntentReviewMonthly:
		text, err := h.review.Monthly(ctx, userID, false)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatAssistantHTML(text)}, nil
	case ai.IntentAnalyticsSummary:
		return h.analyticsView(ctx, userID)
	case ai.IntentNoteCreate:
		if strings.TrimSpace(intent.Title) == "" {
			return dispatchResult{text: "Пример: «запиши заметку идея для Jarvis»"}, nil
		}
		dto, err := h.createNote.Execute(ctx, knowledgeapp.CreateNoteInput{
			UserID: userID, Body: intent.Title, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatNoteCreated(dto)}, err
	case ai.IntentNoteList:
		items, err := h.listNotes.Execute(ctx, knowledgeapp.ListNotesInput{
			UserID: userID, Tag: intent.Target,
		})
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatNotes(items)}, nil
	case ai.IntentNoteSearch:
		if strings.TrimSpace(intent.Title) == "" {
			return dispatchResult{text: "Пример: «найди заметку Jarvis»"}, nil
		}
		items, err := h.searchNotes.Execute(ctx, knowledgeapp.SearchNotesInput{
			UserID: userID, Query: intent.Title,
		})
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatNoteSearchResults(intent.Title, items)}, nil
	case ai.IntentNoteDelete:
		noteID, hint, err := h.resolveNoteToDelete(ctx, userID, intent.Title)
		if err != nil {
			if errors.Is(err, knowledgeapp.ErrNoteNotFound) {
				return dispatchResult{text: FormatNoteNotFound(hint)}, nil
			}
			return dispatchResult{}, err
		}
		dto, err := h.deleteNote.Execute(ctx, knowledgeapp.DeleteNoteInput{
			UserID: userID, NoteID: noteID, Source: events.SourceTelegram,
		})
		if err != nil {
			if errors.Is(err, knowledgeapp.ErrNoteNotFound) {
				return dispatchResult{text: FormatNoteNotFound(hint)}, nil
			}
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatNoteDeleted(dto)}, nil
	case ai.IntentCareerContactCreate:
		if strings.TrimSpace(intent.Title) == "" {
			return dispatchResult{text: "Пример: «добавь контакт Иван — Яндекс»"}, nil
		}
		name, company, role := careerdomain.ParseContactLine(intent.Title)
		dto, err := h.createContact.Execute(ctx, careerapp.CreateContactInput{
			UserID: userID, Name: name, Company: company, Role: role, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatContactCreated(dto)}, err
	case ai.IntentCareerContactList:
		items, err := h.listContacts.Execute(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatContacts(items)}, nil
	case ai.IntentCareerContactSearch:
		if strings.TrimSpace(intent.Title) == "" {
			return dispatchResult{text: "Пример: «найди контакт Иван»"}, nil
		}
		items, err := h.searchContacts.Execute(ctx, careerapp.SearchContactsInput{
			UserID: userID, Query: intent.Title,
		})
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatContactSearchResults(intent.Title, items)}, nil
	case ai.IntentCareerContactDelete:
		contactID, hint, err := h.resolveContactToDelete(ctx, userID, intent.Title)
		if err != nil {
			if errors.Is(err, careerapp.ErrContactNotFound) {
				return dispatchResult{text: FormatContactNotFound(hint)}, nil
			}
			return dispatchResult{}, err
		}
		dto, err := h.deleteContact.Execute(ctx, careerapp.DeleteContactInput{
			UserID: userID, ContactID: contactID, Source: events.SourceTelegram,
		})
		if err != nil {
			if errors.Is(err, careerapp.ErrContactNotFound) {
				return dispatchResult{text: FormatContactNotFound(hint)}, nil
			}
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatContactDeleted(dto)}, nil
	case ai.IntentCareerSkillCreate:
		if strings.TrimSpace(intent.Title) == "" {
			return dispatchResult{text: "Пример: «навык Go senior»"}, nil
		}
		name, level := careerdomain.ParseSkillLine(intent.Title)
		dto, err := h.createSkill.Execute(ctx, careerapp.CreateSkillInput{
			UserID: userID, Name: name, Level: level, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatSkillCreated(dto)}, err
	case ai.IntentCareerSkillList:
		items, err := h.listSkills.Execute(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatSkills(items)}, nil
	case ai.IntentCareerSkillSearch:
		if strings.TrimSpace(intent.Title) == "" {
			return dispatchResult{text: "Пример: «найди навык Go»"}, nil
		}
		items, err := h.searchSkills.Execute(ctx, careerapp.SearchSkillsInput{
			UserID: userID, Query: intent.Title,
		})
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatSkillSearchResults(intent.Title, items)}, nil
	case ai.IntentCareerSkillDelete:
		skillID, hint, err := h.resolveSkillToDelete(ctx, userID, intent.Title)
		if err != nil {
			if errors.Is(err, careerapp.ErrSkillNotFound) {
				return dispatchResult{text: FormatSkillNotFound(hint)}, nil
			}
			return dispatchResult{}, err
		}
		dto, err := h.deleteSkill.Execute(ctx, careerapp.DeleteSkillInput{
			UserID: userID, SkillID: skillID, Source: events.SourceTelegram,
		})
		if err != nil {
			if errors.Is(err, careerapp.ErrSkillNotFound) {
				return dispatchResult{text: FormatSkillNotFound(hint)}, nil
			}
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatSkillDeleted(dto)}, nil
	case ai.IntentSphereCreate:
		if strings.TrimSpace(intent.Title) == "" {
			return h.beginAddSphere(ctx, userID)
		}
		dto, err := h.createSphere.Execute(ctx, spheresapp.CreateSphereInput{
			UserID: userID, Name: intent.Title, Source: events.SourceTelegram,
		})
		if err != nil {
			return dispatchResult{}, err
		}
		settings, serr := h.settingsView(ctx, userID)
		if serr != nil {
			return dispatchResult{text: FormatSphereCreated(dto)}, nil
		}
		return dispatchResult{
			text:   FormatSphereCreated(dto) + "\n\n" + settings.text,
			inline: settings.inline,
		}, nil
	case ai.IntentSphereList:
		return h.settingsView(ctx, userID)
	case ai.IntentSphereUpdate:
		oldName := strings.TrimSpace(intent.Title)
		newName := strings.TrimSpace(intent.Target)
		if oldName == "" || newName == "" {
			return dispatchResult{text: "Пример: «переименуй сферу Карьера в Работа»"}, nil
		}
		sphere, err := h.findSphere.Execute(ctx, userID, oldName)
		if errors.Is(err, spheresdomain.ErrNotFound) {
			return dispatchResult{text: FormatSphereNotFound(oldName)}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		dto, err := h.updateSphere.Execute(ctx, spheresapp.UpdateSphereInput{
			UserID: userID, SphereID: sphere.ID, Name: newName, SortOrder: sphere.SortOrder, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatSphereUpdated(dto)}, err
	case ai.IntentSphereDelete:
		sphereID, hint, err := h.resolveSphereToDelete(ctx, userID, intent.Title)
		if err != nil {
			if errors.Is(err, spheresapp.ErrSphereNotFound) {
				return dispatchResult{text: FormatSphereNotFound(hint)}, nil
			}
			if errors.Is(err, spheresdomain.ErrHasProjects) {
				return dispatchResult{text: "Нельзя удалить сферу с привязанными проектами."}, nil
			}
			return dispatchResult{}, err
		}
		dto, err := h.deleteSphere.Execute(ctx, spheresapp.DeleteSphereInput{
			UserID: userID, SphereID: sphereID, Source: events.SourceTelegram,
		})
		if err != nil {
			if errors.Is(err, spheresapp.ErrSphereNotFound) {
				return dispatchResult{text: FormatSphereNotFound(hint)}, nil
			}
			if errors.Is(err, spheresdomain.ErrHasProjects) {
				return dispatchResult{text: "Нельзя удалить сферу с привязанными проектами."}, nil
			}
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatSphereDeleted(dto)}, nil
	case ai.IntentHealthRecordWeight:
		w, err := rulebased.ParseWeightKg(intent.Title)
		if err != nil {
			return dispatchResult{text: "Пример: «вес 78.5»"}, nil
		}
		dto, err := h.recordWeight.Execute(ctx, healthapp.RecordWeightInput{
			UserID: userID, WeightKg: w, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatWeightRecorded(dto)}, err
	case ai.IntentHealthLatestWeight:
		dto, err := h.latestWeight.Execute(ctx, userID)
		if errors.Is(err, domain.ErrNotFound) {
			return dispatchResult{text: "Записей веса пока нет."}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatLatestWeight(dto)}, nil
	case ai.IntentHealthRecordSteps:
		steps, err := rulebased.ParseStepCount(intent.Title)
		if err != nil {
			return dispatchResult{text: "Пример: «шаги 8000»"}, nil
		}
		dto, err := h.recordSteps.Execute(ctx, healthapp.RecordStepsInput{
			UserID: userID, Steps: steps, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatStepsRecorded(dto)}, err
	case ai.IntentHealthLatestSteps:
		dto, err := h.latestSteps.Execute(ctx, userID)
		if errors.Is(err, domain.ErrNotFound) {
			return dispatchResult{text: "Записей шагов пока нет."}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatLatestSteps(dto)}, nil
	case ai.IntentHealthRecordSleep:
		var mins int32
		if intent.Minute > 0 || intent.Hour > 0 {
			mins = rulebased.HoursMinutesToSleepMinutes(intent.Hour, intent.Minute)
		} else {
			hours, err := rulebased.ParseSleepHours(intent.Title)
			if err != nil {
				return dispatchResult{text: "Пример: «спал 7 часов» или «сон 7.5»"}, nil
			}
			mins = rulebased.SleepHoursToMinutes(hours)
		}
		dto, err := h.recordSleep.Execute(ctx, healthapp.RecordSleepInput{
			UserID: userID, DurationMinutes: mins, Source: events.SourceTelegram,
		})
		return dispatchResult{text: FormatSleepRecorded(dto)}, err
	case ai.IntentHealthLatestSleep:
		dto, err := h.latestSleep.Execute(ctx, userID)
		if errors.Is(err, domain.ErrNotFound) {
			return dispatchResult{text: "Записей сна пока нет."}, nil
		}
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatLatestSleep(dto)}, nil
	default:
		return dispatchResult{text: FormatFallback()}, nil
	}
}

func (h *MessageHandler) resolveReminderToCancel(ctx context.Context, userID ids.UserID, hint string) (uuid.UUID, string, error) {
	items, err := h.listReminders.Execute(ctx, userID)
	if err != nil {
		return uuid.Nil, hint, err
	}
	return matchReminderToCancel(items, hint)
}

// matchReminderToCancel picks the newest pending reminder (empty hint) or the
// first whose message contains hint (case-insensitive).
func matchReminderToCancel(items []notifapp.ReminderDTO, hint string) (uuid.UUID, string, error) {
	hint = strings.TrimSpace(hint)
	if len(items) == 0 {
		return uuid.Nil, hint, notifapp.ErrReminderNotFound
	}
	if hint == "" {
		id, err := uuid.Parse(items[0].ID)
		if err != nil {
			return uuid.Nil, hint, err
		}
		return id, hint, nil
	}
	needle := strings.ToLower(hint)
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Message), needle) {
			id, err := uuid.Parse(item.ID)
			if err != nil {
				return uuid.Nil, hint, err
			}
			return id, hint, nil
		}
	}
	return uuid.Nil, hint, notifapp.ErrReminderNotFound
}

// ensureFutureFireAt kept as a thin wrapper for existing telegram tests.
func ensureFutureFireAt(fireAt, now time.Time) time.Time {
	return rulebased.EnsureFutureFireAt(fireAt, now)
}

func (h *MessageHandler) resolveNoteToDelete(ctx context.Context, userID ids.UserID, hint string) (ids.NoteID, string, error) {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		items, err := h.listNotes.Execute(ctx, knowledgeapp.ListNotesInput{UserID: userID})
		if err != nil {
			return ids.NoteID{}, hint, err
		}
		if len(items) == 0 {
			return ids.NoteID{}, hint, knowledgeapp.ErrNoteNotFound
		}
		return items[0].ID, hint, nil
	}
	items, err := h.searchNotes.Execute(ctx, knowledgeapp.SearchNotesInput{
		UserID: userID, Query: hint,
	})
	if err != nil {
		return ids.NoteID{}, hint, err
	}
	if len(items) == 0 {
		return ids.NoteID{}, hint, knowledgeapp.ErrNoteNotFound
	}
	return items[0].ID, hint, nil
}

func (h *MessageHandler) resolveContactToDelete(ctx context.Context, userID ids.UserID, hint string) (ids.ContactID, string, error) {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		items, err := h.listContacts.Execute(ctx, userID)
		if err != nil {
			return ids.ContactID{}, hint, err
		}
		if len(items) == 0 {
			return ids.ContactID{}, hint, careerapp.ErrContactNotFound
		}
		return items[0].ID, hint, nil
	}
	items, err := h.searchContacts.Execute(ctx, careerapp.SearchContactsInput{
		UserID: userID, Query: hint,
	})
	if err != nil {
		return ids.ContactID{}, hint, err
	}
	if len(items) == 0 {
		return ids.ContactID{}, hint, careerapp.ErrContactNotFound
	}
	return items[0].ID, hint, nil
}

func (h *MessageHandler) resolveSkillToDelete(ctx context.Context, userID ids.UserID, hint string) (ids.SkillID, string, error) {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		items, err := h.listSkills.Execute(ctx, userID)
		if err != nil {
			return ids.SkillID{}, hint, err
		}
		if len(items) == 0 {
			return ids.SkillID{}, hint, careerapp.ErrSkillNotFound
		}
		return items[0].ID, hint, nil
	}
	items, err := h.searchSkills.Execute(ctx, careerapp.SearchSkillsInput{
		UserID: userID, Query: hint,
	})
	if err != nil {
		return ids.SkillID{}, hint, err
	}
	if len(items) == 0 {
		return ids.SkillID{}, hint, careerapp.ErrSkillNotFound
	}
	return items[0].ID, hint, nil
}

func (h *MessageHandler) resolveSphereToDelete(ctx context.Context, userID ids.UserID, hint string) (ids.SphereID, string, error) {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		items, err := h.listSpheres.Execute(ctx, userID)
		if err != nil {
			return ids.SphereID{}, hint, err
		}
		if len(items) == 0 {
			return ids.SphereID{}, hint, spheresapp.ErrSphereNotFound
		}
		return items[len(items)-1].ID, hint, nil
	}
	dto, err := h.findSphere.Execute(ctx, userID, hint)
	if err != nil {
		return ids.SphereID{}, hint, err
	}
	return dto.ID, hint, nil
}

func (h *MessageHandler) projectsBySphereView(ctx context.Context, userID ids.UserID, sphereName string) (dispatchResult, error) {
	sphere, err := h.findSphere.Execute(ctx, userID, sphereName)
	if errors.Is(err, spheresdomain.ErrNotFound) {
		return dispatchResult{text: FormatSphereNotFound(sphereName)}, nil
	}
	if err != nil {
		return dispatchResult{}, err
	}
	items, err := h.listProjects.Execute(ctx, projectsapp.ListProjectsInput{UserID: userID, SphereID: sphere.ID})
	if err != nil {
		return dispatchResult{}, err
	}
	text, inline := FormatProjectsPicker(items)
	if sphereName != "" {
		text = fmt.Sprintf("📁 <b>Проекты сферы %s</b>\n%s", html.EscapeString(sphereName), text)
	}
	return dispatchResult{text: text, inline: inline}, nil
}

func (h *MessageHandler) resolveProjectIDs(ctx context.Context, userID ids.UserID, unit, target string) ([]ids.ProjectID, error) {
	if target == "" {
		return nil, nil
	}
	names := splitAndNames(target)
	out := make([]ids.ProjectID, 0, len(names))
	for _, name := range names {
		project, err := h.findProject.Execute(ctx, userID, name)
		if err != nil {
			return nil, err
		}
		out = append(out, project.ID)
	}
	return out, nil
}

func (h *MessageHandler) resolveSphereIDs(ctx context.Context, userID ids.UserID, unit, target string) ([]ids.SphereID, error) {
	if target != "" {
		names := splitAndNames(target)
		out := make([]ids.SphereID, 0, len(names))
		for _, name := range names {
			sphere, err := h.findSphere.Execute(ctx, userID, name)
			if err != nil {
				return nil, err
			}
			out = append(out, sphere.ID)
		}
		return out, nil
	}
	return h.defaultSphereID(ctx, userID)
}

func (h *MessageHandler) defaultSphereID(ctx context.Context, userID ids.UserID) ([]ids.SphereID, error) {
	candidates := append([]string{}, spheresdomain.DefaultSphereNames...)
	candidates = append(candidates, "карьера", "Деньги")
	for _, name := range candidates {
		sphere, err := h.findSphere.Execute(ctx, userID, name)
		if err == nil {
			return []ids.SphereID{sphere.ID}, nil
		}
		if !errors.Is(err, spheresdomain.ErrNotFound) {
			return nil, err
		}
	}
	spheres, err := h.listSpheres.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(spheres) == 0 {
		return nil, spheresdomain.ErrNotFound
	}
	return []ids.SphereID{spheres[0].ID}, nil
}

func splitAndNames(raw string) []string {
	parts := strings.Split(raw, " и ")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if name := strings.TrimSpace(part); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func atoiIntent(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
