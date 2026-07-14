package rulebased

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type Resolver struct {
	now func() time.Time
}

func NewResolver() *Resolver {
	return &Resolver{now: func() time.Time { return time.Now() }}
}

var patterns = []struct {
	re       *regexp.Regexp
	intent   ai.IntentType
	linkType string
}{
	{regexp.MustCompile(`(?i)^добавь\s+задачу\s+(.+?)\s+для\s+проектов\s+(.+)$`), ai.IntentTaskCreate, "projects"},
	{regexp.MustCompile(`(?i)^добавь\s+задачу\s+(.+?)\s+для\s+проекта\s+(.+)$`), ai.IntentTaskCreate, "project"},
	{regexp.MustCompile(`(?i)^добавь\s+задачу\s+(.+)$`), ai.IntentTaskCreate, ""},
	{regexp.MustCompile(`(?i)^задачи\s+на\s+сегодня`), ai.IntentTaskListToday, ""},
	{regexp.MustCompile(`(?i)^задачи\s+(?:с\s+тегом|по\s+тегу)\s+#?([\p{L}\p{N}_-]+)$`), ai.IntentTaskListByTag, ""},
	{regexp.MustCompile(`(?i)^выполни\s+задачу\s+(.+)$`), ai.IntentTaskComplete, ""},
	{regexp.MustCompile(`(?i)^готово[:\s]+(.+)$`), ai.IntentTaskComplete, ""},
	{regexp.MustCompile(`(?i)^отмени\s+задачу\s+(.+)$`), ai.IntentTaskCancel, ""},
	{regexp.MustCompile(`(?i)^перенеси\s+задачу\s+(.+?)\s+на\s+завтра$`), ai.IntentTaskRescheduleOne, "tomorrow"},
	{regexp.MustCompile(`(?i)^(сколько\s+осталось\s+до\s+проекта|прогресс\s+проекта)(?:\s+(.+))?$`), ai.IntentProjectProgress, ""},
	{regexp.MustCompile(`(?i)^(что\s+сейчас\s+важн|что\s+сейчас\s+самое\s+важн)`), ai.IntentQueryPriorities, ""},
	{regexp.MustCompile(`(?i)^отмени\s+напоминание(?:\s+(.+))?$`), ai.IntentReminderCancel, ""},
	{regexp.MustCompile(`(?i)^напомни\s+(.+)$`), ai.IntentReminderCreate, ""},
	{regexp.MustCompile(`(?i)^работаю\s+до\s+(\d{1,2})(?::(\d{2}))?`), ai.IntentPlanSetAvailability, ""},
	{regexp.MustCompile(`(?i)^сегодня\s+полный\s+завал`), ai.IntentPlanTriage, ""},
	{regexp.MustCompile(`(?i)^перенеси\s+задачи`), ai.IntentTaskReschedule, ""},
	{regexp.MustCompile(`(?i)^утренний\s+обзор\s+в\s+(\d{1,2})(?::(\d{2}))?`), ai.IntentSettingsMorning, ""},
	{regexp.MustCompile(`(?i)^вечерний\s+обзор\s+в\s+(\d{1,2})(?::(\d{2}))?`), ai.IntentSettingsEvening, ""},
	{regexp.MustCompile(`(?i)^тихие\s+часы\s+с\s+(\d{1,2})(?::(\d{2}))?\s+до\s+(\d{1,2})(?::(\d{2}))?`), ai.IntentSettingsQuietHours, ""},
	{regexp.MustCompile(`(?i)^пришёл\s+заказ\s+на\s+(.+)$`), ai.IntentFinanceIncome, ""},
	{regexp.MustCompile(`(?i)^(?:пришёл|пришла|получил|получила|заработал)\s+(.+)$`), ai.IntentFinanceIncome, ""},
	{regexp.MustCompile(`(?i)^доход\s+(.+)$`), ai.IntentFinanceIncome, ""},
	{regexp.MustCompile(`(?i)^запиши\s+доход\s+(.+)$`), ai.IntentFinanceIncome, ""},
	{regexp.MustCompile(`(?i)^потратил\s+(.+?)\s+на\s+(.+)$`), ai.IntentFinanceExpense, ""},
	{regexp.MustCompile(`(?i)^расход\s+(.+)$`), ai.IntentFinanceExpense, ""},
	{regexp.MustCompile(`(?i)^покажи\s+долги`), ai.IntentFinanceListDebts, ""},
	{regexp.MustCompile(`(?i)^мои\s+долги`), ai.IntentFinanceListDebts, ""},
	{regexp.MustCompile(`(?i)^финансы\s+за\s+месяц`), ai.IntentFinanceCashFlow, ""},
	{regexp.MustCompile(`(?i)^(?:заплатил|оплатил|вернул|отдал)\s+([\d\s]+(?:тысяч|тыс|к)?)\s+(.+)$`), ai.IntentFinancePayDebt, ""},
	{regexp.MustCompile(`(?i)^долг\s+([\d\s]+(?:тысяч|тыс|к)?)\s+(.+)$`), ai.IntentFinanceCreateDebt, ""},
	{regexp.MustCompile(`(?i)^добавь\s+привычку\s+(.+)$`), ai.IntentHabitCreate, ""},
	{regexp.MustCompile(`(?i)^отметь\s+привычку\s+(.+)$`), ai.IntentHabitTrack, ""},
	{regexp.MustCompile(`(?i)^привычка\s+(.+)\s+готово$`), ai.IntentHabitTrack, ""},
	{regexp.MustCompile(`(?i)^мои\s+привычки`), ai.IntentHabitList, ""},
	{regexp.MustCompile(`(?i)^привычки\s+на\s+сегодня`), ai.IntentHabitList, ""},
	{regexp.MustCompile(`(?i)^добавь\s+(?:встречу|событие)\s+(.+?)\s+завтра\s+в\s+(\d{1,2})(?::(\d{2}))?`), ai.IntentCalendarCreate, "завтра"},
	{regexp.MustCompile(`(?i)^добавь\s+(?:встречу|событие)\s+(.+?)\s+сегодня\s+в\s+(\d{1,2})(?::(\d{2}))?`), ai.IntentCalendarCreate, "сегодня"},
	{regexp.MustCompile(`(?i)^встреча\s+(.+?)\s+завтра\s+в\s+(\d{1,2})(?::(\d{2}))?`), ai.IntentCalendarCreate, "завтра"},
	{regexp.MustCompile(`(?i)^(календарь|события)\s+на\s+сегодня`), ai.IntentCalendarListToday, ""},
	{regexp.MustCompile(`(?i)^задачи\s+проекта\s+(.+)$`), ai.IntentProjectTasks, ""},
	{regexp.MustCompile(`(?i)^(?:архивируй|закрой)\s+проект\s+(.+)$`), ai.IntentProjectArchive, ""},
	{regexp.MustCompile(`(?i)^добавь\s+проект\s+(.+?)\s+в\s+сферах\s+(.+)$`), ai.IntentProjectCreate, "spheres"},
	{regexp.MustCompile(`(?i)^добавь\s+проект\s+(.+?)\s+в\s+сфере\s+(.+)$`), ai.IntentProjectCreate, "sphere"},
	{regexp.MustCompile(`(?i)^добавь\s+проект\s+(.+)$`), ai.IntentProjectCreate, ""},
	{regexp.MustCompile(`(?i)^проекты\s+сферы\s+(.+)$`), ai.IntentProjectList, "sphere"},
	{regexp.MustCompile(`(?i)^мои\s+проекты`), ai.IntentProjectList, ""},
	{regexp.MustCompile(`(?i)^список\s+проектов`), ai.IntentProjectList, ""},
	{regexp.MustCompile(`(?i)^недельный\s+обзор`), ai.IntentReviewWeekly, ""},
	{regexp.MustCompile(`(?i)^месячный\s+обзор`), ai.IntentReviewMonthly, ""},
	{regexp.MustCompile(`(?i)^(?:аналитика|моя\s+статистика|статистика\s+за\s+месяц|продуктивность)`), ai.IntentAnalyticsSummary, ""},
	{regexp.MustCompile(`(?i)^запиши\s+заметку\s+(.+)$`), ai.IntentNoteCreate, ""},
	{regexp.MustCompile(`(?i)^заметка[:\s]+(.+)$`), ai.IntentNoteCreate, ""},
	{regexp.MustCompile(`(?i)^заметки\s+(?:с\s+тегом\s+|#)([^\s]+)$`), ai.IntentNoteList, "tag"},
	{regexp.MustCompile(`(?i)^(?:мои\s+заметки|последние\s+заметки)`), ai.IntentNoteList, ""},
	{regexp.MustCompile(`(?i)^(?:найди\s+заметку|поиск\s+заметок|найди\s+в\s+заметках)\s+(.+)$`), ai.IntentNoteSearch, ""},
	{regexp.MustCompile(`(?i)^удали\s+заметку(?:\s+(.+))?$`), ai.IntentNoteDelete, ""},
	{regexp.MustCompile(`(?i)^(?:добавь|запиши)\s+контакт\s+(.+)$`), ai.IntentCareerContactCreate, ""},
	{regexp.MustCompile(`(?i)^контакт[:\s]+(.+)$`), ai.IntentCareerContactCreate, ""},
	{regexp.MustCompile(`(?i)^(?:мои\s+контакты|контакты|список\s+контактов)$`), ai.IntentCareerContactList, ""},
	{regexp.MustCompile(`(?i)^(?:найди\s+контакт|поиск\s+контактов)\s+(.+)$`), ai.IntentCareerContactSearch, ""},
	{regexp.MustCompile(`(?i)^удали\s+контакт(?:\s+(.+))?$`), ai.IntentCareerContactDelete, ""},
	{regexp.MustCompile(`(?i)^(?:добавь|запиши)\s+навык\s+(.+)$`), ai.IntentCareerSkillCreate, ""},
	{regexp.MustCompile(`(?i)^навык[:\s]+(.+)$`), ai.IntentCareerSkillCreate, ""},
	{regexp.MustCompile(`(?i)^(?:мои\s+навыки|навыки|список\s+навыков)$`), ai.IntentCareerSkillList, ""},
	{regexp.MustCompile(`(?i)^(?:найди\s+навык|поиск\s+навыков)\s+(.+)$`), ai.IntentCareerSkillSearch, ""},
	{regexp.MustCompile(`(?i)^удали\s+навык(?:\s+(.+))?$`), ai.IntentCareerSkillDelete, ""},
	{regexp.MustCompile(`(?i)^(?:добавь|запиши)\s+сферу\s+(.+)$`), ai.IntentSphereCreate, ""},
	{regexp.MustCompile(`(?i)^(?:мои\s+сферы|сферы|список\s+сфер)$`), ai.IntentSphereList, ""},
	{regexp.MustCompile(`(?i)^переименуй\s+сферу\s+(.+?)\s+в\s+(.+)$`), ai.IntentSphereUpdate, ""},
	{regexp.MustCompile(`(?i)^удали\s+сферу(?:\s+(.+))?$`), ai.IntentSphereDelete, ""},
	{regexp.MustCompile(`(?i)^запиши\s+вес\s+([\d]+(?:[.,]\d+)?)\s*(?:кг)?$`), ai.IntentHealthRecordWeight, ""},
	{regexp.MustCompile(`(?i)^вес\s+([\d]+(?:[.,]\d+)?)\s*(?:кг)?$`), ai.IntentHealthRecordWeight, ""},
	{regexp.MustCompile(`(?i)^(?:мой\s+вес|последний\s+вес|текущий\s+вес)$`), ai.IntentHealthLatestWeight, ""},
	{regexp.MustCompile(`(?i)^запиши\s+шаги\s+([\d\s]+)$`), ai.IntentHealthRecordSteps, ""},
	{regexp.MustCompile(`(?i)^шаги\s+([\d\s]+)$`), ai.IntentHealthRecordSteps, ""},
	{regexp.MustCompile(`(?i)^(?:мои\s+шаги|последние\s+шаги|сколько\s+шагов)$`), ai.IntentHealthLatestSteps, ""},
	{regexp.MustCompile(`(?i)^запиши\s+сон\s+([\d]+(?:[.,]\d+)?)\s*(?:ч(?:ас(?:ов|а)?)?)?$`), ai.IntentHealthRecordSleep, ""},
	{regexp.MustCompile(`(?i)^(?:спал|сон)\s+([\d]+(?:[.,]\d+)?)\s*(?:ч(?:ас(?:ов|а)?)?)?$`), ai.IntentHealthRecordSleep, ""},
	{regexp.MustCompile(`(?i)^(?:спал|сон)\s+([\d]+)\s*(?:ч(?:ас(?:ов|а)?)?)\s+([\d]+)\s*(?:мин(?:ут(?:ы)?)?|м)?$`), ai.IntentHealthRecordSleep, "hm"},
	{regexp.MustCompile(`(?i)^(?:мой\s+сон|последний\s+сон|сколько\s+спал)$`), ai.IntentHealthLatestSleep, ""},
}

func (r *Resolver) Resolve(_ context.Context, input ai.ResolveInput) (ai.ResolvedIntent, error) {
	text := strings.TrimSpace(input.Text)
	for _, p := range patterns {
		if m := p.re.FindStringSubmatch(text); m != nil {
			intent := ai.ResolvedIntent{Type: p.intent, Confidence: 0.9, Unit: p.linkType}
			switch p.intent {
			case ai.IntentTaskCreate, ai.IntentTaskComplete, ai.IntentTaskCancel, ai.IntentTaskRescheduleOne, ai.IntentTaskListByTag, ai.IntentHabitCreate, ai.IntentHabitTrack, ai.IntentProjectCreate, ai.IntentProjectTasks, ai.IntentProjectArchive, ai.IntentProjectProgress, ai.IntentNoteCreate, ai.IntentNoteSearch, ai.IntentCareerContactCreate, ai.IntentCareerContactSearch, ai.IntentCareerSkillCreate, ai.IntentCareerSkillSearch, ai.IntentSphereCreate:
				intent.Title = strings.TrimSpace(m[1])
				if p.intent == ai.IntentTaskCreate && len(m) > 2 && m[2] != "" {
					intent.Target = strings.TrimSpace(m[2])
				}
				if p.intent == ai.IntentProjectCreate && len(m) > 2 && m[2] != "" {
					intent.Target = strings.TrimSpace(m[2])
				}
				if p.intent == ai.IntentProjectProgress && len(m) > 2 && m[2] != "" {
					intent.Title = strings.TrimSpace(m[2])
				}
			case ai.IntentCalendarCreate:
				intent.Title = strings.TrimSpace(m[1])
				intent.Target = p.linkType
				intent.Hour = atoiDefault(m[2], 10)
				if len(m) > 3 && m[3] != "" {
					intent.Minute = atoiDefault(m[3], 0)
				}
			case ai.IntentReminderCreate:
				intent.Message = strings.TrimSpace(m[1])
				intent.TimeText = extractTimeHint(intent.Message)
			case ai.IntentReminderCancel:
				if len(m) > 1 && m[1] != "" {
					intent.Message = strings.TrimSpace(m[1])
				}
			case ai.IntentNoteDelete:
				if len(m) > 1 && m[1] != "" {
					intent.Title = strings.TrimSpace(m[1])
				}
			case ai.IntentCareerContactDelete:
				if len(m) > 1 && m[1] != "" {
					intent.Title = strings.TrimSpace(m[1])
				}
			case ai.IntentCareerSkillDelete:
				if len(m) > 1 && m[1] != "" {
					intent.Title = strings.TrimSpace(m[1])
				}
			case ai.IntentSphereUpdate:
				intent.Title = strings.TrimSpace(m[1])
				intent.Target = strings.TrimSpace(m[2])
			case ai.IntentSphereDelete:
				if len(m) > 1 && m[1] != "" {
					intent.Title = strings.TrimSpace(m[1])
				}
			case ai.IntentProjectList:
				if p.linkType == "sphere" && len(m) > 1 {
					intent.Target = strings.TrimSpace(m[1])
				}
			case ai.IntentNoteList:
				if p.linkType == "tag" && len(m) > 1 {
					intent.Target = strings.TrimSpace(m[1])
				}
			case ai.IntentPlanSetAvailability:
				intent.Hour = atoiDefault(m[1], 0)
				if len(m) > 2 {
					intent.Minute = atoiDefault(m[2], 0)
				}
			case ai.IntentSettingsMorning, ai.IntentSettingsEvening:
				intent.Hour = atoiDefault(m[1], 0)
				if len(m) > 2 {
					intent.Minute = atoiDefault(m[2], 0)
				}
			case ai.IntentSettingsQuietHours:
				intent.Hour = atoiDefault(m[1], 0)
				if len(m) > 2 && m[2] != "" {
					intent.Minute = atoiDefault(m[2], 0)
				}
				if len(m) > 3 {
					intent.Target = m[3]
				}
				if len(m) > 4 && m[4] != "" {
					intent.Unit = m[4]
				}
			case ai.IntentFinanceIncome:
				raw := strings.TrimSpace(m[1])
				if cents, err := ParseRublesAmount(raw); err == nil {
					intent.AmountCents = cents
				}
				intent.Title = incomeDescription(text)
				intent.Currency = "RUB"
			case ai.IntentFinanceExpense:
				raw := strings.TrimSpace(m[1])
				if cents, err := ParseRublesAmount(raw); err == nil {
					intent.AmountCents = cents
				}
				if len(m) > 2 {
					intent.Title = strings.TrimSpace(m[2])
				} else {
					intent.Title = "Прочее"
				}
				intent.Currency = "RUB"
			case ai.IntentFinanceCreateDebt:
				raw := strings.TrimSpace(m[1])
				if cents, err := ParseRublesAmount(raw); err == nil {
					intent.AmountCents = cents
				}
				intent.Target = strings.TrimSpace(m[2])
				intent.Currency = "RUB"
			case ai.IntentFinancePayDebt:
				raw := strings.TrimSpace(m[1])
				if cents, err := ParseRublesAmount(raw); err == nil {
					intent.AmountCents = cents
				}
				intent.Target = strings.TrimSpace(m[2])
				intent.Currency = "RUB"
			case ai.IntentHealthRecordWeight:
				intent.Title = strings.TrimSpace(m[1])
			case ai.IntentHealthRecordSteps:
				intent.Title = strings.TrimSpace(m[1])
			case ai.IntentHealthRecordSleep:
				if p.linkType == "hm" {
					intent.Hour = atoiDefault(m[1], 0)
					intent.Minute = atoiDefault(m[2], 0)
				} else {
					intent.Title = strings.TrimSpace(m[1])
				}
			}
			return intent, nil
		}
	}
	return ai.ResolvedIntent{Type: ai.IntentUnknown, Confidence: 0}, nil
}

func extractTimeHint(s string) string {
	lower := strings.ToLower(s)
	for _, hint := range []string{"вечером", "утром", "завтра", "через"} {
		if strings.Contains(lower, hint) {
			return hint
		}
	}
	return "вечером"
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}
