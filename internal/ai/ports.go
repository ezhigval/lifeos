package ai

import "context"

type IntentType string

const (
	IntentTaskCreate          IntentType = "task.create"
	IntentTaskListToday       IntentType = "task.list_today"
	IntentTaskComplete        IntentType = "task.complete"
	IntentTaskCancel          IntentType = "task.cancel"
	IntentTaskRescheduleOne   IntentType = "task.reschedule_one"
	IntentTaskListByTag       IntentType = "task.list_by_tag"
	IntentProjectProgress     IntentType = "project.progress"
	IntentQueryPriorities     IntentType = "query.priorities"
	IntentReminderCreate      IntentType = "reminder.create"
	IntentReminderCancel      IntentType = "reminder.cancel"
	IntentPlanSetAvailability IntentType = "plan.set_availability"
	IntentPlanTriage          IntentType = "plan.triage"
	IntentTaskReschedule      IntentType = "task.reschedule"
	IntentSettingsMorning     IntentType = "settings.morning_review"
	IntentSettingsEvening     IntentType = "settings.evening_review"
	IntentSettingsQuietHours  IntentType = "settings.quiet_hours"
	IntentFinanceIncome       IntentType = "finance.income"
	IntentFinanceExpense      IntentType = "finance.expense"
	IntentFinanceListDebts    IntentType = "finance.list_debts"
	IntentFinanceCashFlow     IntentType = "finance.cash_flow"
	IntentFinanceCreateDebt   IntentType = "finance.create_debt"
	IntentFinancePayDebt      IntentType = "finance.pay_debt"
	IntentHabitCreate         IntentType = "habit.create"
	IntentHabitTrack          IntentType = "habit.track"
	IntentHabitList           IntentType = "habit.list"
	IntentProjectCreate       IntentType = "project.create"
	IntentProjectList         IntentType = "project.list"
	IntentProjectTasks        IntentType = "project.tasks"
	IntentProjectArchive      IntentType = "project.archive"
	IntentCalendarCreate      IntentType = "calendar.create"
	IntentCalendarListToday   IntentType = "calendar.list_today"
	IntentReviewWeekly        IntentType = "review.weekly"
	IntentReviewMonthly       IntentType = "review.monthly"
	IntentAnalyticsSummary    IntentType = "analytics.summary"
	IntentNoteCreate          IntentType = "note.create"
	IntentNoteList            IntentType = "note.list"
	IntentNoteSearch          IntentType = "note.search"
	IntentNoteDelete          IntentType = "note.delete"
	IntentHealthRecordWeight  IntentType = "health.record_weight"
	IntentHealthLatestWeight  IntentType = "health.latest_weight"
	IntentHealthRecordSteps   IntentType = "health.record_steps"
	IntentHealthLatestSteps   IntentType = "health.latest_steps"
	IntentHealthRecordSleep   IntentType = "health.record_sleep"
	IntentHealthLatestSleep   IntentType = "health.latest_sleep"
	IntentCareerContactCreate IntentType = "career.contact_create"
	IntentCareerContactList   IntentType = "career.contact_list"
	IntentCareerContactSearch IntentType = "career.contact_search"
	IntentCareerContactDelete IntentType = "career.contact_delete"
	IntentCareerSkillCreate   IntentType = "career.skill_create"
	IntentCareerSkillList     IntentType = "career.skill_list"
	IntentCareerSkillSearch   IntentType = "career.skill_search"
	IntentCareerSkillDelete   IntentType = "career.skill_delete"
	IntentSphereCreate        IntentType = "sphere.create"
	IntentSphereList          IntentType = "sphere.list"
	IntentSphereUpdate        IntentType = "sphere.update"
	IntentSphereDelete        IntentType = "sphere.delete"
	IntentUnknown             IntentType = "unknown"
)

type ResolveInput struct {
	Text     string
	Language string
}

type ResolvedIntent struct {
	Type        IntentType
	Title       string
	Message     string
	Target      string
	Unit        string
	TimeText    string
	Hour        int
	Minute      int
	AmountCents int64
	Currency    string
	Confidence  float64
}

type IntentResolver interface {
	Resolve(ctx context.Context, input ResolveInput) (ResolvedIntent, error)
}

type SummaryRequest struct {
	Tasks    []string
	Projects []string
}

type Assistant interface {
	Summarize(ctx context.Context, req SummaryRequest) (string, error)
}
