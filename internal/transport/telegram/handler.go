package telegram

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
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
	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	identitydomain "github.com/valentinezhov/lifeos/internal/identity/domain"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	planapp "github.com/valentinezhov/lifeos/internal/planning/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	projectsdomain "github.com/valentinezhov/lifeos/internal/projects/domain"
	"github.com/valentinezhov/lifeos/internal/query"
	settingsapp "github.com/valentinezhov/lifeos/internal/settings/app"
	settingsdomain "github.com/valentinezhov/lifeos/internal/settings/domain"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	spheresdomain "github.com/valentinezhov/lifeos/internal/spheres/domain"
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
	log              *slog.Logger
	client           *Client
	screen           *Screen
	sessions         *tginfra.Sessions
	ensureUser       *identityapp.EnsureUserByTelegram
	processed        *tginfra.ProcessedUpdates
	resolver         ai.IntentResolver
	createTask       *tasksapp.CreateTask
	completeTask     *tasksapp.CompleteTask
	completeByTitle  *tasksapp.CompleteTaskByTitle
	projectProg      *projectsapp.GetProjectProgress
	listToday        *tasksapp.ListTasksToday
	priorities       *query.GetTopPriorities
	reminder         *notifapp.ScheduleReminder
	listReminders    *notifapp.ListReminders
	cancelReminder   *notifapp.CancelReminder
	setAvail         *planapp.SetDayAvailability
	triage           *planapp.TriageOverloadedDay
	reschedule       *planapp.RescheduleTasks
	recordIncome     *financeapp.RecordIncome
	recordExpense    *financeapp.RecordExpense
	createDebt       *financeapp.CreateDebt
	payDebt          *financeapp.PayDebt
	listDebts        *financeapp.ListDebts
	cashFlow         *financeapp.CashFlowSummary
	createHabit      *habitsapp.CreateHabit
	trackHabit       *habitsapp.TrackHabit
	listHabits       *habitsapp.ListHabitsToday
	createProject    *projectsapp.CreateProject
	findProject      *projectsapp.FindProjectByName
	listProjects     *projectsapp.ListProjects
	listProjectTasks *tasksapp.ListTasksByProject
	archiveProject   *projectsapp.ArchiveProject
	createEvent      *calendarapp.CreateEvent
	listCalendar     *calendarapp.ListEventsToday
	review           *query.Review
	analytics        *query.GetProductivitySummary
	createNote       *knowledgeapp.CreateNote
	listNotes        *knowledgeapp.ListNotes
	searchNotes      *knowledgeapp.SearchNotes
	deleteNote       *knowledgeapp.DeleteNote
	createContact    *careerapp.CreateContact
	listContacts     *careerapp.ListContacts
	searchContacts   *careerapp.SearchContacts
	deleteContact    *careerapp.DeleteContact
	createSkill      *careerapp.CreateSkill
	listSkills       *careerapp.ListSkills
	searchSkills     *careerapp.SearchSkills
	deleteSkill      *careerapp.DeleteSkill
	createSphere     *spheresapp.CreateSphere
	listSpheres      *spheresapp.ListSpheres
	updateSphere     *spheresapp.UpdateSphere
	deleteSphere     *spheresapp.DeleteSphere
	findSphere       *spheresapp.FindSphereByName
	recordWeight     *healthapp.RecordWeight
	latestWeight     *healthapp.GetLatestWeight
	recordSteps      *healthapp.RecordSteps
	latestSteps      *healthapp.GetLatestSteps
	recordSleep      *healthapp.RecordSleep
	latestSleep      *healthapp.GetLatestSleep
	updateMorning    *settingsapp.UpdateMorningReview
	updateEvening    *settingsapp.UpdateEveningReview
	updateQuiet      *settingsapp.UpdateQuietHours
	tzReader         interface {
		Timezone(ctx context.Context, userID ids.UserID) (string, error)
	}
	deleteUser      *identityapp.DeleteUser
	adminTelegramID int64
	miniAppURL      string
}

type Deps struct {
	Log              *slog.Logger
	Client           *Client
	Sessions         *tginfra.Sessions
	EnsureUser       *identityapp.EnsureUserByTelegram
	Processed        *tginfra.ProcessedUpdates
	Resolver         ai.IntentResolver
	CreateTask       *tasksapp.CreateTask
	CompleteTask     *tasksapp.CompleteTask
	CompleteByTitle  *tasksapp.CompleteTaskByTitle
	ProjectProg      *projectsapp.GetProjectProgress
	ListToday        *tasksapp.ListTasksToday
	Priorities       *query.GetTopPriorities
	Reminder         *notifapp.ScheduleReminder
	ListReminders    *notifapp.ListReminders
	CancelReminder   *notifapp.CancelReminder
	SetAvail         *planapp.SetDayAvailability
	Triage           *planapp.TriageOverloadedDay
	Reschedule       *planapp.RescheduleTasks
	RecordIncome     *financeapp.RecordIncome
	RecordExpense    *financeapp.RecordExpense
	CreateDebt       *financeapp.CreateDebt
	PayDebt          *financeapp.PayDebt
	ListDebts        *financeapp.ListDebts
	CashFlow         *financeapp.CashFlowSummary
	CreateHabit      *habitsapp.CreateHabit
	TrackHabit       *habitsapp.TrackHabit
	ListHabits       *habitsapp.ListHabitsToday
	CreateProject    *projectsapp.CreateProject
	FindProject      *projectsapp.FindProjectByName
	ListProjects     *projectsapp.ListProjects
	ListProjectTasks *tasksapp.ListTasksByProject
	ArchiveProject   *projectsapp.ArchiveProject
	CreateEvent      *calendarapp.CreateEvent
	ListCalendar     *calendarapp.ListEventsToday
	Review           *query.Review
	Analytics        *query.GetProductivitySummary
	CreateNote       *knowledgeapp.CreateNote
	ListNotes        *knowledgeapp.ListNotes
	SearchNotes      *knowledgeapp.SearchNotes
	DeleteNote       *knowledgeapp.DeleteNote
	CreateContact    *careerapp.CreateContact
	ListContacts     *careerapp.ListContacts
	SearchContacts   *careerapp.SearchContacts
	DeleteContact    *careerapp.DeleteContact
	CreateSkill      *careerapp.CreateSkill
	ListSkills       *careerapp.ListSkills
	SearchSkills     *careerapp.SearchSkills
	DeleteSkill      *careerapp.DeleteSkill
	CreateSphere     *spheresapp.CreateSphere
	ListSpheres      *spheresapp.ListSpheres
	UpdateSphere     *spheresapp.UpdateSphere
	DeleteSphere     *spheresapp.DeleteSphere
	FindSphere       *spheresapp.FindSphereByName
	RecordWeight     *healthapp.RecordWeight
	LatestWeight     *healthapp.GetLatestWeight
	RecordSteps      *healthapp.RecordSteps
	LatestSteps      *healthapp.GetLatestSteps
	RecordSleep      *healthapp.RecordSleep
	LatestSleep      *healthapp.GetLatestSleep
	UpdateMorning    *settingsapp.UpdateMorningReview
	UpdateEvening    *settingsapp.UpdateEveningReview
	UpdateQuiet      *settingsapp.UpdateQuietHours
	TZReader         interface {
		Timezone(ctx context.Context, userID ids.UserID) (string, error)
	}
	DeleteUser      *identityapp.DeleteUser
	AdminTelegramID int64
	MiniAppURL      string
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
		projectProg: d.ProjectProg, listToday: d.ListToday,
		priorities: d.Priorities, reminder: d.Reminder, listReminders: d.ListReminders, cancelReminder: d.CancelReminder, setAvail: d.SetAvail,
		triage: d.Triage, reschedule: d.Reschedule,
		recordIncome: d.RecordIncome, recordExpense: d.RecordExpense,
		createDebt: d.CreateDebt, payDebt: d.PayDebt, listDebts: d.ListDebts, cashFlow: d.CashFlow,
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
	}
}

func (h *MessageHandler) HandleUpdate(ctx context.Context, update Update) error {
	if update.CallbackQuery != nil {
		return h.handleCallback(ctx, update)
	}
	if update.Message == nil || update.Message.Text == "" {
		return nil
	}
	seen, err := h.processed.Seen(ctx, update.UpdateID)
	if err != nil {
		return err
	}
	if seen {
		return nil
	}

	user, err := h.ensureUser.Execute(ctx, identityapp.EnsureUserInput{
		TelegramID:  update.Message.From.ID,
		DisplayName: FormatDisplayName(update.Message.From),
	})
	if err != nil {
		return fmt.Errorf("resolve user: %w", err)
	}

	chatID := update.Message.Chat.ID
	userMsgID := update.Message.MessageID
	in := classifyInput(update.Message.Text)

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
		}
	}

	if isCancelText(text) && sess.State != tginfra.StateIdle {
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
		_ = h.sessions.SetState(ctx, userID, tginfra.StateIdle, map[string]any{"defer_tasks": strs})
	}
	return h.screen.Show(ctx, userID, chatID, out.text, out.inline, MainReplyKeyboard(h.miniAppURL))
}

func (h *MessageHandler) runAction(ctx context.Context, userID ids.UserID, action string) (dispatchResult, error) {
	switch action {
	case ActionAddTask:
		if err := h.sessions.SetState(ctx, userID, tginfra.StateAwaitTaskTitle, h.basePayload(ctx, userID)); err != nil {
			return dispatchResult{}, err
		}
	case ActionAddProject:
		if err := h.sessions.SetState(ctx, userID, tginfra.StateAwaitProjectTitle, h.basePayload(ctx, userID)); err != nil {
			return dispatchResult{}, err
		}
	default:
		if err := h.sessions.SetStateOnly(ctx, userID, tginfra.StateIdle); err != nil {
			return dispatchResult{}, err
		}
	}

	switch action {
	case ActionHome:
		items, err := h.listToday.Execute(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatHomeSummary(len(items)), inline: InlineHomeActions()}, nil
	case ActionTasksToday:
		return h.tasksTodayView(ctx, userID)
	case ActionPriorities:
		items, err := h.priorities.Execute(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatPriorities(items)}, nil
	case ActionAddTask:
		return dispatchResult{text: FormatPromptTaskTitle()}, nil
	case ActionProjectProgress:
		p, err := h.projectProg.Execute(ctx, userID, ids.ProjectID{})
		if err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatProjectProgress(p)}, nil
	case ActionAddProject:
		return dispatchResult{text: FormatPromptProjectTitle()}, nil
	case ActionTriage:
		msg, low, err := h.triage.Propose(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		out := dispatchResult{text: msg, deferTasks: low}
		if len(low) > 0 {
			out.inline = [][]InlineButton{{{Text: "Перенести low", CallbackData: CBTriageDefer}}}
		}
		return out, nil
	case ActionHabits:
		return h.habitsTodayView(ctx, userID)
	case ActionProjects:
		return h.projectsPickerView(ctx, userID)
	case ActionCalendar:
		return h.calendarTodayView(ctx, userID)
	case ActionAnalytics:
		return h.analyticsView(ctx, userID)
	case ActionSettings:
		return h.settingsView(ctx, userID)
	case ActionMiniApp:
		if h.miniAppURL == "" {
			return dispatchResult{text: "Mini App ещё не настроен (нет LIFEOS_MINIAPP_URL)."}, nil
		}
		return dispatchResult{text: "📱 Mini App: " + html.EscapeString(h.miniAppURL) + "\nНажми кнопку <b>📱 Mini App</b> на клавиатуре, чтобы открыть."}, nil
	default:
		return dispatchResult{text: FormatFallback()}, nil
	}
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
	replyKBSetKey     = "reply_kb_set"
	replyKBVersionKey = "reply_kb_ver"
	// Bump when MainReplyKeyboard layout or attach strategy changes.
	replyKBVersion = 4
)

func replyKeyboardInstalled(payload map[string]any) bool {
	if payload == nil || payload[replyKBSetKey] != true {
		return false
	}
	ver, ok := payloadInt64(payload, replyKBVersionKey)
	return ok && ver == replyKBVersion
}

func (h *MessageHandler) resetReplyKeyboardFlag(ctx context.Context, userID ids.UserID) error {
	sess, err := h.sessions.Get(ctx, userID)
	if err != nil {
		return err
	}
	if sess.StatePayload == nil {
		return nil
	}
	delete(sess.StatePayload, replyKBSetKey)
	delete(sess.StatePayload, replyKBVersionKey)
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

func (h *MessageHandler) completeTaskByID(ctx context.Context, userID ids.UserID, rawID string) (dispatchResult, error) {
	taskID, err := ids.ParseTaskID(rawID)
	if err != nil {
		return dispatchResult{}, err
	}
	dto, err := h.completeTask.Execute(ctx, tasksapp.CompleteTaskInput{
		UserID: userID, TaskID: taskID, Source: events.SourceTelegram,
	})
	if err != nil {
		return dispatchResult{}, err
	}
	if view, ok := h.viewProjectID(ctx, userID); ok {
		project, err := h.findProjectByID(ctx, userID, view)
		if err != nil {
			return dispatchResult{text: FormatTaskCompleted(dto)}, nil
		}
		tasks, err := h.projectTasksView(ctx, userID, view, project.Name)
		if err != nil {
			return dispatchResult{text: FormatTaskCompleted(dto)}, nil
		}
		return dispatchResult{text: FormatTaskCompleted(dto) + "\n\n" + tasks.text, inline: tasks.inline}, nil
	}
	today, err := h.tasksTodayView(ctx, userID)
	if err != nil {
		return dispatchResult{text: FormatTaskCompleted(dto)}, nil
	}
	return dispatchResult{text: FormatTaskCompleted(dto) + "\n\n" + today.text, inline: today.inline}, nil
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

func (h *MessageHandler) dispatchIntent(ctx context.Context, userID ids.UserID, intent ai.ResolvedIntent) (dispatchResult, error) {
	switch intent.Type {
	case ai.IntentTaskCreate:
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
		tz, err := h.tzReader.Timezone(ctx, userID)
		if err != nil {
			return dispatchResult{}, err
		}
		fireAt := rulebased.ParseFireAt(time.Now().UTC(), tz, intent.TimeText)
		if err := h.reminder.Execute(ctx, notifapp.ScheduleReminderInput{
			UserID: userID, Message: intent.Message, FireAt: fireAt,
		}); err != nil {
			return dispatchResult{}, err
		}
		return dispatchResult{text: FormatReminderScheduled(fireAt.Format("15:04"))}, nil
	case ai.IntentReminderCancel:
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
		return dispatchResult{text: FormatReminderCancelled(dto.Message, dto.FireAt.Format("15:04 02.01"))}, nil
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
		return dispatchResult{text: text}, err
	case ai.IntentReviewMonthly:
		text, err := h.review.Monthly(ctx, userID, false)
		return dispatchResult{text: text}, err
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
	if len(items) == 0 {
		return uuid.Nil, hint, notifapp.ErrReminderNotFound
	}
	hint = strings.TrimSpace(hint)
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
	for _, name := range []string{"Карьера GO", "Деньги", "карьера"} {
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
