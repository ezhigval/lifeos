package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
	calendarapp "github.com/valentinezhov/lifeos/internal/calendar/app"
	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	identitydomain "github.com/valentinezhov/lifeos/internal/identity/domain"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	planapp "github.com/valentinezhov/lifeos/internal/planning/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	projectsdomain "github.com/valentinezhov/lifeos/internal/projects/domain"
	"github.com/valentinezhov/lifeos/internal/query"
	settingsapp "github.com/valentinezhov/lifeos/internal/settings/app"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
	tginfra "github.com/valentinezhov/lifeos/internal/transport/telegram/infra"
	"golang.org/x/sync/errgroup"
)

type dispatchResult struct {
	text       string
	inline     [][]InlineButton
	deferTasks []ids.TaskID
}

type MessageHandler struct {
	log               *slog.Logger
	client            *Client
	screen            *Screen
	sessions          *tginfra.Sessions
	ensureUser        *identityapp.EnsureUserByTelegram
	processed         *tginfra.ProcessedUpdates
	resolver          ai.IntentResolver
	createTask        *tasksapp.CreateTask
	completeTask      *tasksapp.CompleteTask
	completeByTitle   *tasksapp.CompleteTaskByTitle
	cancelByTitle     *tasksapp.CancelTaskByTitle
	rescheduleByTitle *tasksapp.RescheduleTaskByTitle
	cancelTask        *tasksapp.CancelTask
	listByTag         *tasksapp.ListTasksByTag
	projectProg       *projectsapp.GetProjectProgress
	listToday         *tasksapp.ListTasksToday
	priorities        *query.GetTopPriorities
	reminder          *notifapp.ScheduleReminder
	listReminders     *notifapp.ListReminders
	cancelReminder    *notifapp.CancelReminder
	setAvail          *planapp.SetDayAvailability
	triage            *planapp.TriageOverloadedDay
	reschedule        *planapp.RescheduleTasks
	recordIncome      *financeapp.RecordIncome
	recordExpense     *financeapp.RecordExpense
	createDebt        *financeapp.CreateDebt
	payDebt           *financeapp.PayDebt
	listDebts         *financeapp.ListDebts
	cashFlow          *financeapp.CashFlowSummary
	listFinancePlan   *financeapp.ListFinancePlan
	createPlanned     *financeapp.CreatePlannedCashflow
	createHabit       *habitsapp.CreateHabit
	trackHabit        *habitsapp.TrackHabit
	listHabits        *habitsapp.ListHabitsToday
	createProject     *projectsapp.CreateProject
	findProject       *projectsapp.FindProjectByName
	listProjects      *projectsapp.ListProjects
	listProjectTasks  *tasksapp.ListTasksByProject
	archiveProject    *projectsapp.ArchiveProject
	createEvent       *calendarapp.CreateEvent
	listCalendar      *calendarapp.ListEventsToday
	review            *query.Review
	analytics         *query.GetProductivitySummary
	createNote        *knowledgeapp.CreateNote
	listNotes         *knowledgeapp.ListNotes
	searchNotes       *knowledgeapp.SearchNotes
	deleteNote        *knowledgeapp.DeleteNote
	createContact     *careerapp.CreateContact
	listContacts      *careerapp.ListContacts
	searchContacts    *careerapp.SearchContacts
	deleteContact     *careerapp.DeleteContact
	createSkill       *careerapp.CreateSkill
	listSkills        *careerapp.ListSkills
	searchSkills      *careerapp.SearchSkills
	deleteSkill       *careerapp.DeleteSkill
	createSphere      *spheresapp.CreateSphere
	listSpheres       *spheresapp.ListSpheres
	updateSphere      *spheresapp.UpdateSphere
	deleteSphere      *spheresapp.DeleteSphere
	findSphere        *spheresapp.FindSphereByName
	recordWeight      *healthapp.RecordWeight
	latestWeight      *healthapp.GetLatestWeight
	recordSteps       *healthapp.RecordSteps
	latestSteps       *healthapp.GetLatestSteps
	recordSleep       *healthapp.RecordSleep
	latestSleep       *healthapp.GetLatestSleep
	updateMorning     *settingsapp.UpdateMorningReview
	updateEvening     *settingsapp.UpdateEveningReview
	updateQuiet       *settingsapp.UpdateQuietHours
	tzReader          interface {
		Timezone(ctx context.Context, userID ids.UserID) (string, error)
	}
	deleteUser      *identityapp.DeleteUser
	adminTelegramID int64
	miniAppURL      string
	agent           *AgentBridge
	stt             ai.SpeechToText
}

type Deps struct {
	Log               *slog.Logger
	Client            *Client
	Sessions          *tginfra.Sessions
	EnsureUser        *identityapp.EnsureUserByTelegram
	Processed         *tginfra.ProcessedUpdates
	Resolver          ai.IntentResolver
	SpeechToText      ai.SpeechToText
	CreateTask        *tasksapp.CreateTask
	CompleteTask      *tasksapp.CompleteTask
	CompleteByTitle   *tasksapp.CompleteTaskByTitle
	CancelByTitle     *tasksapp.CancelTaskByTitle
	RescheduleByTitle *tasksapp.RescheduleTaskByTitle
	CancelTask        *tasksapp.CancelTask
	ListByTag         *tasksapp.ListTasksByTag
	ProjectProg       *projectsapp.GetProjectProgress
	ListToday         *tasksapp.ListTasksToday
	Priorities        *query.GetTopPriorities
	Reminder          *notifapp.ScheduleReminder
	ListReminders     *notifapp.ListReminders
	CancelReminder    *notifapp.CancelReminder
	SetAvail          *planapp.SetDayAvailability
	Triage            *planapp.TriageOverloadedDay
	Reschedule        *planapp.RescheduleTasks
	RecordIncome      *financeapp.RecordIncome
	RecordExpense     *financeapp.RecordExpense
	CreateDebt        *financeapp.CreateDebt
	PayDebt           *financeapp.PayDebt
	ListDebts         *financeapp.ListDebts
	CashFlow          *financeapp.CashFlowSummary
	ListFinancePlan   *financeapp.ListFinancePlan
	CreatePlanned     *financeapp.CreatePlannedCashflow
	CreateHabit       *habitsapp.CreateHabit
	TrackHabit        *habitsapp.TrackHabit
	ListHabits        *habitsapp.ListHabitsToday
	CreateProject     *projectsapp.CreateProject
	FindProject       *projectsapp.FindProjectByName
	ListProjects      *projectsapp.ListProjects
	ListProjectTasks  *tasksapp.ListTasksByProject
	ArchiveProject    *projectsapp.ArchiveProject
	CreateEvent       *calendarapp.CreateEvent
	ListCalendar      *calendarapp.ListEventsToday
	Review            *query.Review
	Analytics         *query.GetProductivitySummary
	CreateNote        *knowledgeapp.CreateNote
	ListNotes         *knowledgeapp.ListNotes
	SearchNotes       *knowledgeapp.SearchNotes
	DeleteNote        *knowledgeapp.DeleteNote
	CreateContact     *careerapp.CreateContact
	ListContacts      *careerapp.ListContacts
	SearchContacts    *careerapp.SearchContacts
	DeleteContact     *careerapp.DeleteContact
	CreateSkill       *careerapp.CreateSkill
	ListSkills        *careerapp.ListSkills
	SearchSkills      *careerapp.SearchSkills
	DeleteSkill       *careerapp.DeleteSkill
	CreateSphere      *spheresapp.CreateSphere
	ListSpheres       *spheresapp.ListSpheres
	UpdateSphere      *spheresapp.UpdateSphere
	DeleteSphere      *spheresapp.DeleteSphere
	FindSphere        *spheresapp.FindSphereByName
	RecordWeight      *healthapp.RecordWeight
	LatestWeight      *healthapp.GetLatestWeight
	RecordSteps       *healthapp.RecordSteps
	LatestSteps       *healthapp.GetLatestSteps
	RecordSleep       *healthapp.RecordSleep
	LatestSleep       *healthapp.GetLatestSleep
	UpdateMorning     *settingsapp.UpdateMorningReview
	UpdateEvening     *settingsapp.UpdateEveningReview
	UpdateQuiet       *settingsapp.UpdateQuietHours
	TZReader          interface {
		Timezone(ctx context.Context, userID ids.UserID) (string, error)
	}
	DeleteUser      *identityapp.DeleteUser
	AdminTelegramID int64
	MiniAppURL      string
	Agent           *AgentBridge
}

func NewHandler(d Deps) *MessageHandler {
	if d.Resolver == nil {
		d.Resolver = rulebased.NewResolver()
	}
	screen := NewScreen(d.Client, d.Sessions)
	return &MessageHandler{
		log: d.Log, client: d.Client, screen: screen, sessions: d.Sessions,
		ensureUser: d.EnsureUser, processed: d.Processed, resolver: d.Resolver,
		createTask: d.CreateTask, completeTask: d.CompleteTask, completeByTitle: d.CompleteByTitle,
		cancelByTitle: d.CancelByTitle, rescheduleByTitle: d.RescheduleByTitle,
		cancelTask: d.CancelTask, listByTag: d.ListByTag,
		projectProg: d.ProjectProg, listToday: d.ListToday,
		priorities: d.Priorities, reminder: d.Reminder, listReminders: d.ListReminders, cancelReminder: d.CancelReminder, setAvail: d.SetAvail,
		triage: d.Triage, reschedule: d.Reschedule,
		recordIncome: d.RecordIncome, recordExpense: d.RecordExpense,
		createDebt: d.CreateDebt, payDebt: d.PayDebt, listDebts: d.ListDebts, cashFlow: d.CashFlow,
		listFinancePlan: d.ListFinancePlan, createPlanned: d.CreatePlanned,
		createHabit: d.CreateHabit, trackHabit: d.TrackHabit, listHabits: d.ListHabits,
		createProject: d.CreateProject, findProject: d.FindProject, listProjects: d.ListProjects,
		listProjectTasks: d.ListProjectTasks, archiveProject: d.ArchiveProject, createEvent: d.CreateEvent, listCalendar: d.ListCalendar,
		review: d.Review, analytics: d.Analytics,
		createNote: d.CreateNote, listNotes: d.ListNotes, searchNotes: d.SearchNotes, deleteNote: d.DeleteNote,
		createContact: d.CreateContact, listContacts: d.ListContacts, searchContacts: d.SearchContacts, deleteContact: d.DeleteContact,
		createSkill: d.CreateSkill, listSkills: d.ListSkills, searchSkills: d.SearchSkills, deleteSkill: d.DeleteSkill,
		createSphere: d.CreateSphere, listSpheres: d.ListSpheres, updateSphere: d.UpdateSphere,
		deleteSphere: d.DeleteSphere, findSphere: d.FindSphere,
		recordWeight: d.RecordWeight, latestWeight: d.LatestWeight,
		recordSteps: d.RecordSteps, latestSteps: d.LatestSteps,
		recordSleep: d.RecordSleep, latestSleep: d.LatestSleep,
		updateMorning: d.UpdateMorning, updateEvening: d.UpdateEvening, updateQuiet: d.UpdateQuiet,
		tzReader:   d.TZReader,
		deleteUser: d.DeleteUser, adminTelegramID: d.AdminTelegramID,
		miniAppURL: strings.TrimSpace(d.MiniAppURL),
		agent:      d.Agent,
		stt:        d.SpeechToText,
	}
}

func (h *MessageHandler) HandleUpdate(ctx context.Context, update Update) error {
	if update.CallbackQuery != nil {
		return h.handleCallback(ctx, update)
	}
	if update.Message == nil {
		return nil
	}

	seen, err := h.processed.Seen(ctx, update.UpdateID)
	if err != nil {
		return err
	}
	if seen {
		return nil
	}

	chatID := update.Message.Chat.ID
	userMsgID := update.Message.MessageID

	text, resolveErr := h.resolveIncomingText(ctx, chatID, update.Message)
	if resolveErr != nil {
		user, uerr := h.ensureUser.Execute(ctx, identityapp.EnsureUserInput{
			TelegramID:  update.Message.From.ID,
			DisplayName: FormatDisplayName(update.Message.From),
		})
		if uerr != nil {
			return fmt.Errorf("resolve user: %w", uerr)
		}
		out := dispatchResult{text: formatMediaResolveError(resolveErr)}
		if err := h.present(ctx, user.ID, chatID, out); err != nil {
			return err
		}
		_ = h.client.DeleteMessage(ctx, chatID, userMsgID)
		return h.processed.Mark(ctx, update.UpdateID)
	}
	if strings.TrimSpace(text) == "" {
		// Unsupported / empty media with nothing to say — ignore silently.
		return nil
	}

	user, err := h.ensureUser.Execute(ctx, identityapp.EnsureUserInput{
		TelegramID:  update.Message.From.ID,
		DisplayName: FormatDisplayName(update.Message.From),
	})
	if err != nil {
		return fmt.Errorf("resolve user: %w", err)
	}

	in := classifyInput(text)

	out, err := h.dispatchNormalized(ctx, user, update.Message.From, chatID, userMsgID, in)
	if err != nil {
		out = dispatchResult{text: "Ошибка: " + err.Error()}
	}
	// /clear already presented a fresh home during wipe.
	if in.IsCommand() && in.Command == CmdClear && err == nil && out.text == "" && len(out.inline) == 0 {
		return h.processed.Mark(ctx, update.UpdateID)
	}
	// Self-delete recreates the user row with a new id — rebind before present.
	if fresh, eerr := h.ensureUser.Execute(ctx, identityapp.EnsureUserInput{
		TelegramID:  update.Message.From.ID,
		DisplayName: FormatDisplayName(update.Message.From),
	}); eerr == nil {
		user = fresh
	}
	if err := h.present(ctx, user.ID, chatID, out); err != nil {
		return err
	}
	if err := h.client.DeleteMessage(ctx, chatID, userMsgID); err != nil {
		h.log.Debug("delete user message failed", "chat_id", chatID, "message_id", userMsgID, "error", err)
	}
	return h.processed.Mark(ctx, update.UpdateID)
}

func (h *MessageHandler) dispatchNormalized(
	ctx context.Context,
	user identitydomain.User,
	from User,
	chatID, userMsgID int64,
	in NormalizedInput,
) (dispatchResult, error) {
	switch {
	case in.IsCommand():
		return h.handleCommand(ctx, user, from, chatID, userMsgID, in)
	case in.IsKeyboard():
		return h.handleKeyboard(ctx, user.ID, in.Action)
	case in.IsText():
		return h.handleFreeText(ctx, user, from, chatID, userMsgID, in.Text)
	default:
		return dispatchResult{text: FormatFallback()}, nil
	}
}

func (h *MessageHandler) handleCommand(
	ctx context.Context,
	user identitydomain.User,
	from User,
	chatID, userMsgID int64,
	in NormalizedInput,
) (dispatchResult, error) {
	switch in.Command {
	case CmdClear:
		if err := h.clearChat(ctx, user.ID, chatID, userMsgID); err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{}, nil
	case CmdDelete:
		return h.beginDeleteUser(ctx, user.ID, from.ID, from.Username, in.Raw)
	case CmdStart:
		_ = h.resetReplyKeyboardFlag(ctx, user.ID)
		if action, ok := commandToAction(in.Command); ok {
			return h.runAction(ctx, user.ID, action)
		}
	case CmdCancel:
		return h.handleFreeText(ctx, user, from, chatID, userMsgID, TextCancel)
	}
	if action, ok := commandToAction(in.Command); ok {
		return h.runAction(ctx, user.ID, action)
	}
	return dispatchResult{text: "Неизвестная команда: " + in.Command}, nil
}

func (h *MessageHandler) handleKeyboard(ctx context.Context, userID ids.UserID, action string) (dispatchResult, error) {
	return h.runAction(ctx, userID, action)
}

func (h *MessageHandler) handleFreeText(
	ctx context.Context,
	user identitydomain.User,
	from User,
	chatID, userMsgID int64,
	text string,
) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, user.ID)
	if err != nil {
		return dispatchResult{}, err
	}

	if _, ok := payloadInt64(sess.StatePayload, PayloadPendingDeleteTG); ok {
		switch {
		case isDeleteConfirmText(text):
			return h.confirmPendingDelete(ctx, user.ID, from.ID, chatID, userMsgID)
		case isCancelText(text):
			return h.cancelPendingDelete(ctx, user.ID)
		default:
			// Do not fall through to intents while a wipe is armed ("да" must stay explicit).
			return dispatchResult{text: "Сначала подтверди удаление («confirm» / кнопка) или отмени («отмена»)."}, nil
		}
	}

	if isCancelText(text) {
		if sess.State == tginfra.StateIdle {
			// Idle cancel must not fall through to the intent resolver ("отмена" ≠ task cancel).
			return dispatchResult{text: "Нечего отменять."}, nil
		}
		if err := h.sessions.SetState(ctx, user.ID, tginfra.StateIdle, h.basePayload(ctx, user.ID)); err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: "Отменено."}, nil
	}

	switch sess.State {
	case tginfra.StateAwaitTaskTitle:
		return h.onTaskTitleEntered(ctx, user.ID, text)
	case tginfra.StateAwaitProjectTitle:
		return h.onProjectTitleEntered(ctx, user.ID, text)
	case tginfra.StateAwaitSphereName:
		return h.onSphereNameEntered(ctx, user.ID, text)
	case tginfra.StateAwaitTaskProjects, tginfra.StateAwaitProjectSpheres:
		return dispatchResult{text: "Выбери варианты кнопками на экране или напиши «отмена»."}, nil
	case tginfra.StateAwaitAgentTurn:
		return h.runAgentDialogue(ctx, user.ID, text, sess)
	}

	// Known commands: rule-based first (fast, deterministic). Agent only for unknown / chat.
	primary := rulebased.NewResolver()
	if intent, err := primary.Resolve(ctx, ai.ResolveInput{Text: text, Language: "ru"}); err == nil && intent.Type != ai.IntentUnknown {
		return h.dispatchIntent(ctx, user.ID, intent)
	}

	if h.agent != nil && h.agent.Agent != nil {
		return h.runAgentDialogue(ctx, user.ID, text, sess)
	}

	intent, err := h.resolver.Resolve(ctx, ai.ResolveInput{Text: text, Language: "ru"})
	if err != nil {
		return dispatchResult{}, err
	}
	return h.dispatchIntent(ctx, user.ID, intent)
}

func (h *MessageHandler) handleCallback(ctx context.Context, update Update) error {
	_ = h.client.AnswerCallback(ctx, update.CallbackQuery.ID)

	user, err := h.ensureUser.Execute(ctx, identityapp.EnsureUserInput{
		TelegramID:  update.CallbackQuery.From.ID,
		DisplayName: FormatDisplayName(update.CallbackQuery.From),
	})
	if err != nil {
		return err
	}

	chatID := update.CallbackQuery.Message.Chat.ID
	data := update.CallbackQuery.Data

	var out dispatchResult
	switch {
	case strings.HasPrefix(data, CBActionPrefix):
		out, err = h.runAction(ctx, user.ID, strings.TrimPrefix(data, CBActionPrefix))
	case data == CBTriageDefer:
		out, err = h.applyTriageDefer(ctx, user.ID)
	case strings.HasPrefix(data, CBTaskDone):
		out, err = h.completeTaskByID(ctx, user.ID, strings.TrimPrefix(data, CBTaskDone))
	case strings.HasPrefix(data, CBTaskCancel):
		out, err = h.cancelTaskByID(ctx, user.ID, strings.TrimPrefix(data, CBTaskCancel))
	case strings.HasPrefix(data, CBHabitTrack):
		out, err = h.trackHabitByID(ctx, user.ID, strings.TrimPrefix(data, CBHabitTrack))
	case strings.HasPrefix(data, CBProjectView):
		out, err = h.projectTasksByID(ctx, user.ID, strings.TrimPrefix(data, CBProjectView))
	case data == CBSettingsAdd:
		out, err = h.beginAddSphere(ctx, user.ID)
	case strings.HasPrefix(data, CBSettingsDel):
		out, err = h.deleteSphereByID(ctx, user.ID, strings.TrimPrefix(data, CBSettingsDel))
	case data == CBDraftSphereOK:
		out, err = h.confirmDraftProject(ctx, user.ID)
	case strings.HasPrefix(data, CBDraftSphere):
		out, err = h.toggleDraftSphere(ctx, user.ID, strings.TrimPrefix(data, CBDraftSphere))
	case data == CBDraftProjectOK:
		out, err = h.confirmDraftTaskWithSelection(ctx, user.ID)
	case data == CBDraftProjectSkip:
		out, err = h.confirmDraftTaskSkip(ctx, user.ID)
	case strings.HasPrefix(data, CBDraftProject):
		out, err = h.toggleDraftProject(ctx, user.ID, strings.TrimPrefix(data, CBDraftProject))
	case data == CBDeleteOK:
		out, err = h.confirmPendingDelete(ctx, user.ID, update.CallbackQuery.From.ID, chatID, update.CallbackQuery.Message.MessageID)
		if err == nil {
			// Self-delete recreates the user with a new id.
			if fresh, eerr := h.ensureUser.Execute(ctx, identityapp.EnsureUserInput{
				TelegramID:  update.CallbackQuery.From.ID,
				DisplayName: FormatDisplayName(update.CallbackQuery.From),
			}); eerr == nil {
				user = fresh
			}
		}
	case data == CBDeleteCancel:
		out, err = h.cancelPendingDelete(ctx, user.ID)
	default:
		return nil
	}
	if err != nil {
		out = dispatchResult{text: "Ошибка: " + err.Error()}
	}
	return h.present(ctx, user.ID, chatID, out)
}

func (h *MessageHandler) present(ctx context.Context, userID ids.UserID, chatID int64, out dispatchResult) error {
	if len(out.deferTasks) > 0 {
		strs := make([]string, len(out.deferTasks))
		for i, id := range out.deferTasks {
			strs[i] = id.String()
		}
		// Preserve reply_kb_* / view_project_id so triage does not force Mini App keyboard reinstall.
		payload := h.basePayload(ctx, userID)
		payload["defer_tasks"] = strs
		_ = h.sessions.SetState(ctx, userID, tginfra.StateIdle, payload)
	}
	return h.screen.Show(ctx, userID, chatID, out.text, out.inline, MainReplyKeyboard(h.miniAppURL), h.miniAppURL)
}

func (h *MessageHandler) settingsView(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	items, err := h.listSpheres.Execute(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	text, inline := FormatSettingsSpheres(items)
	text = FormatSectionHeader("⚙️ Настройки", text) + "\n\nМожно также писать: «добавь сферу …», «переименуй сферу X в Y»"
	return dispatchResult{text: text, inline: inline}, nil
}

const (
	replyKBSetKey        = "reply_kb_set"
	replyKBVersionKey    = "reply_kb_ver"
	replyKBMiniAppKey    = "reply_kb_miniapp"
	replyKBMiniAppURLKey = "reply_kb_miniapp_url"
	// Bump when MainReplyKeyboard layout or URL-tracking strategy changes.
	// v7: Mini App reply row is plain text; launch via inline web_app (initData).
	replyKBVersion = 7
)

func replyKeyboardHasMiniApp(rows [][]ReplyButton) bool {
	for _, row := range rows {
		for _, btn := range row {
			if btn.Text == MenuMiniApp {
				return true
			}
		}
	}
	return false
}

func replyKeyboardInstalled(payload map[string]any, wantMiniApp bool, wantURL string) bool {
	if payload == nil || payload[replyKBSetKey] != true {
		return false
	}
	ver, ok := payloadInt64(payload, replyKBVersionKey)
	if !ok || ver != replyKBVersion {
		return false
	}
	hadMiniApp, _ := payload[replyKBMiniAppKey].(bool)
	if hadMiniApp != wantMiniApp {
		return false
	}
	if wantMiniApp {
		hadURL, _ := payload[replyKBMiniAppURLKey].(string)
		if strings.TrimSpace(hadURL) != strings.TrimSpace(wantURL) {
			return false // tunnel URL rotated → force keyboard reinstall
		}
	}
	return true
}

func (h *MessageHandler) resetReplyKeyboardFlag(ctx context.Context, userID ids.UserID) error {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return err
	}
	if sess.StatePayload == nil {
		sess.StatePayload = map[string]any{}
	}
	delete(sess.StatePayload, replyKBSetKey)
	delete(sess.StatePayload, replyKBVersionKey)
	delete(sess.StatePayload, replyKBMiniAppKey)
	delete(sess.StatePayload, replyKBMiniAppURLKey)
	// Clear dashboard so /start resends a message that re-attaches the reply keyboard.
	sess.DashboardMessageID = 0
	return h.sessions.Save(ctx, sess)
}

const clearChatMessageWindow = 200

// clearChat resets conversation UI to a fresh start without wiping domain data.
// It best-effort deletes recent chat messages (Telegram private chats use sequential IDs).
func (h *MessageHandler) clearChat(ctx context.Context, userID ids.UserID, chatID, lastMsgID int64) error {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return err
	}

	high := lastMsgID
	if sess.DashboardMessageID > high {
		high = sess.DashboardMessageID
	}

	if err := h.sessions.Reset(ctx, userID, chatID); err != nil {
		return fmt.Errorf("reset session: %w", err)
	}

	low := high - clearChatMessageWindow + 1
	if low < 1 {
		low = 1
	}
	h.deleteMessageRange(ctx, chatID, low, high)

	out, err := h.runAction(ctx, userID, ActionHome)
	if err != nil {
		return err
	}
	if out.text != "" {
		out.text = "♻️ Переписка сброшена. Твои данные на месте.\n\n" + out.text
	} else {
		out.text = "♻️ Переписка сброшена. Твои данные на месте."
	}
	return h.present(ctx, userID, chatID, out)
}

func (h *MessageHandler) deleteMessageRange(ctx context.Context, chatID, low, high int64) {
	if high < low {
		return
	}
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	for id := high; id >= low; id-- {
		id := id
		g.Go(func() error {
			_ = h.client.DeleteMessage(gctx, chatID, id)
			return nil
		})
	}
	_ = g.Wait()
}

func (h *MessageHandler) habitsTodayView(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	items, err := h.listHabits.Execute(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	text, inline := FormatHabitsWithActions(items)
	return dispatchResult{text: text, inline: inline}, nil
}

func (h *MessageHandler) projectsPickerView(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	items, err := h.listProjects.Execute(ctx, projectsapp.ListProjectsInput{UserID: userID})
	if err != nil {
		return dispatchResult{}, err
	}
	text, inline := FormatProjectsPicker(items)
	return dispatchResult{text: text, inline: inline}, nil
}

func (h *MessageHandler) projectTasksByID(ctx context.Context, userID ids.UserID, rawID string) (dispatchResult, error) {
	projectID, err := ids.ParseProjectID(rawID)
	if err != nil {
		return dispatchResult{}, err
	}
	project, err := h.findProjectByID(ctx, userID, projectID)
	if err != nil {
		return dispatchResult{}, err
	}
	return h.projectTasksView(ctx, userID, project.ID, project.Name)
}

func (h *MessageHandler) projectTasksByName(ctx context.Context, userID ids.UserID, name string) (dispatchResult, error) {
	project, err := h.findProject.Execute(ctx, userID, name)
	if errors.Is(err, projectsdomain.ErrNotFound) {
		return dispatchResult{text: FormatProjectNotFound(name)}, nil
	}
	if err != nil {
		return dispatchResult{}, err
	}
	return h.projectTasksView(ctx, userID, project.ID, project.Name)
}

func (h *MessageHandler) findProjectByID(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) (projectsapp.ProjectDTO, error) {
	items, err := h.listProjects.Execute(ctx, projectsapp.ListProjectsInput{UserID: userID})
	if err != nil {
		return projectsapp.ProjectDTO{}, err
	}
	for _, item := range items {
		if item.ID == projectID {
			return item, nil
		}
	}
	return projectsapp.ProjectDTO{}, projectsdomain.ErrNotFound
}

func (h *MessageHandler) projectTasksView(ctx context.Context, userID ids.UserID, projectID ids.ProjectID, projectName string) (dispatchResult, error) {
	items, err := h.listProjectTasks.Execute(ctx, userID, projectID)
	if err != nil {
		return dispatchResult{}, err
	}
	_ = h.setViewProject(ctx, userID, projectID)
	text, inline := FormatProjectTasksWithActions(projectName, items)
	return dispatchResult{text: text, inline: inline}, nil
}

func (h *MessageHandler) setViewProject(ctx context.Context, userID ids.UserID, projectID ids.ProjectID) error {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return err
	}
	if sess.StatePayload == nil {
		sess.StatePayload = map[string]any{}
	}
	sess.StatePayload["view_project_id"] = projectID.String()
	sess.State = tginfra.StateIdle
	return h.sessions.Save(ctx, sess)
}

func (h *MessageHandler) viewProjectID(ctx context.Context, userID ids.UserID) (ids.ProjectID, bool) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return ids.ProjectID{}, false
	}
	raw, ok := sess.StatePayload["view_project_id"].(string)
	if !ok || raw == "" {
		return ids.ProjectID{}, false
	}
	id, err := ids.ParseProjectID(raw)
	if err != nil {
		return ids.ProjectID{}, false
	}
	return id, true
}

func (h *MessageHandler) calendarTodayView(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	items, err := h.listCalendar.Execute(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	tz, err := h.tzReader.Timezone(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	return dispatchResult{text: FormatSectionHeader("📆 Календарь", FormatCalendarToday(items, tz))}, nil
}

func (h *MessageHandler) analyticsView(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	summary, err := h.analytics.Execute(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	return dispatchResult{text: FormatSectionHeader("📊 Статистика", query.FormatProductivitySummary(summary))}, nil
}

func (h *MessageHandler) trackHabitByID(ctx context.Context, userID ids.UserID, rawID string) (dispatchResult, error) {
	habitID, err := ids.ParseHabitID(rawID)
	if err != nil {
		return dispatchResult{}, err
	}
	result, err := h.trackHabit.ExecuteByID(ctx, userID, habitID, events.SourceTelegram)
	if err != nil {
		return dispatchResult{}, err
	}
	today, err := h.habitsTodayView(ctx, userID)
	if err != nil {
		return dispatchResult{text: FormatHabitTracked(result)}, nil
	}
	return dispatchResult{text: FormatHabitTracked(result) + "\n\n" + today.text, inline: today.inline}, nil
}

func (h *MessageHandler) tasksTodayView(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	items, err := h.listToday.Execute(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	text, inline := FormatTasksTodayWithActions(items)
	return dispatchResult{text: text, inline: inline}, nil
}

func (h *MessageHandler) cancelTaskByID(ctx context.Context, userID ids.UserID, rawID string) (dispatchResult, error) {
	taskID, err := ids.ParseTaskID(rawID)
	if err != nil {
		return dispatchResult{}, err
	}
	dto, err := h.cancelTask.Execute(ctx, tasksapp.CancelTaskInput{
		UserID: userID, TaskID: taskID, Source: events.SourceTelegram,
	})
	if err != nil {
		msg := formatTaskLifecycleCallbackError(err, true)
		if msg == "" {
			return dispatchResult{}, err
		}
		return h.prependTaskListRefresh(ctx, userID, msg)
	}
	return h.prependTaskListRefresh(ctx, userID, FormatTaskCancelled(dto))
}

func (h *MessageHandler) completeTaskByID(ctx context.Context, userID ids.UserID, rawID string) (dispatchResult, error) {
	taskID, err := ids.ParseTaskID(rawID)
	if err != nil {
		return dispatchResult{}, err
	}
	dto, err := h.completeTask.Execute(ctx, tasksapp.CompleteTaskInput{
		UserID: userID, TaskID: taskID, Source: events.SourceTelegram,
	})
	if err != nil {
		msg := formatTaskLifecycleCallbackError(err, false)
		if msg == "" {
			return dispatchResult{}, err
		}
		return h.prependTaskListRefresh(ctx, userID, msg)
	}
	return h.prependTaskListRefresh(ctx, userID, FormatTaskCompleted(dto))
}

// formatTaskLifecycleCallbackError maps stale ✕/✓ taps to RU copy; empty means rethrow.
func formatTaskLifecycleCallbackError(err error, cancel bool) string {
	switch {
	case errors.Is(err, tasksapp.ErrTaskNotFound):
		return "Задача не найдена."
	case cancel && errors.Is(err, taskdomain.ErrAlreadyCancelled):
		return "Задача уже отменена."
	case cancel && errors.Is(err, taskdomain.ErrCannotCancelDone):
		return "Выполненную задачу нельзя отменить."
	case !cancel && errors.Is(err, taskdomain.ErrCannotComplete):
		return "Отменённую задачу нельзя выполнить."
	default:
		return ""
	}
}

func (h *MessageHandler) prependTaskListRefresh(ctx context.Context, userID ids.UserID, head string) (dispatchResult, error) {
	if view, ok := h.viewProjectID(ctx, userID); ok {
		project, err := h.findProjectByID(ctx, userID, view)
		if err != nil {
			return dispatchResult{text: head}, nil
		}
		tasks, err := h.projectTasksView(ctx, userID, view, project.Name)
		if err != nil {
			return dispatchResult{text: head}, nil
		}
		return dispatchResult{text: head + "\n\n" + tasks.text, inline: tasks.inline}, nil
	}
	today, err := h.tasksTodayView(ctx, userID)
	if err != nil {
		return dispatchResult{text: head}, nil
	}
	return dispatchResult{text: head + "\n\n" + today.text, inline: today.inline}, nil
}

func (h *MessageHandler) applyTriageDefer(ctx context.Context, userID ids.UserID) (dispatchResult, error) {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return dispatchResult{}, err
	}
	raw, ok := sess.StatePayload["defer_tasks"]
	if !ok {
		return dispatchResult{text: "Нечего переносить."}, nil
	}
	taskIDs := parseTaskIDs(raw)
	n, err := h.triage.ApplyDefer(ctx, userID, taskIDs)
	if err != nil {
		return dispatchResult{}, err
	}
	_ = h.sessions.SetState(ctx, userID, tginfra.StateIdle, h.basePayload(ctx, userID))
	return dispatchResult{text: fmt.Sprintf("Перенесено задач: %d", n)}, nil
}

func parseTaskIDs(raw any) []ids.TaskID {
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]ids.TaskID, 0, len(arr))
	for _, v := range arr {
		s, ok := v.(string)
		if !ok {
			continue
		}
		id, err := ids.ParseTaskID(s)
		if err == nil {
			out = append(out, id)
		}
	}
	return out
}
