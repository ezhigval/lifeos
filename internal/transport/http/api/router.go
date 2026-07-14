package api

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	calendarapp "github.com/valentinezhov/lifeos/internal/calendar/app"
	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/auth"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	"github.com/valentinezhov/lifeos/internal/query"
	settingsapp "github.com/valentinezhov/lifeos/internal/settings/app"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
)

type Deps struct {
	Log              *slog.Logger
	APIKey           string
	BotToken         string
	WebAppAuthTTL    time.Duration
	Tokens           *auth.TokenService
	GetUser          *identityapp.GetUserByTelegram
	EnsureUser       *identityapp.EnsureUserByTelegram
	ListToday        *tasksapp.ListTasksToday
	CreateTask       *tasksapp.CreateTask
	Complete         *tasksapp.CompleteTask
	CancelTask       *tasksapp.CancelTask
	EditTask         *tasksapp.EditTask
	RescheduleTask   *tasksapp.RescheduleTask
	ListByTag        *tasksapp.ListTasksByTag
	GetTask          *tasksapp.GetTask
	UpdateTask       *tasksapp.UpdateTask
	ArchiveTask      *tasksapp.ArchiveTask
	DeleteTask       *tasksapp.DeleteTask
	ProjectProg      *projectsapp.GetProjectProgress
	Review           *query.Review
	Priorities       *query.GetTopPriorities
	Analytics        *query.GetProductivitySummary
	RecordIncome     *financeapp.RecordIncome
	RecordExpense    *financeapp.RecordExpense
	ListDebts        *financeapp.ListDebts
	CreateDebt       *financeapp.CreateDebt
	PayDebt          *financeapp.PayDebt
	CashFlow         *financeapp.CashFlowSummary
	ListHabits       *habitsapp.ListHabitsToday
	CreateHabit      *habitsapp.CreateHabit
	TrackHabit       *habitsapp.TrackHabit
	ScheduleReminder *notifapp.ScheduleReminder
	ListReminders    *notifapp.ListReminders
	CancelReminder   *notifapp.CancelReminder
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
	RecordWeight     *healthapp.RecordWeight
	GetLatestWeight  *healthapp.GetLatestWeight
	ListWeights      *healthapp.ListWeights
	RecordSteps      *healthapp.RecordSteps
	GetLatestSteps   *healthapp.GetLatestSteps
	ListSteps        *healthapp.ListSteps
	RecordSleep      *healthapp.RecordSleep
	GetLatestSleep   *healthapp.GetLatestSleep
	ListSleep        *healthapp.ListSleep
	ListCalendar     *calendarapp.ListEventsToday
	CreateEvent      *calendarapp.CreateEvent
	ListProjects     *projectsapp.ListProjects
	CreateProject    *projectsapp.CreateProject
	ListProjectTasks *tasksapp.ListTasksByProject
	ArchiveProject   *projectsapp.ArchiveProject
	GetSettings      *settingsapp.GetSettings
	UpdateMorning    *settingsapp.UpdateMorningReview
	UpdateEvening    *settingsapp.UpdateEveningReview
	UpdateQuiet      *settingsapp.UpdateQuietHours
}

type Router struct {
	deps Deps
}

func NewRouter(deps Deps) *Router {
	return &Router{deps: deps}
}

func (rt *Router) Mount(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/token", rt.issueToken)
		r.Post("/auth/telegram-webapp", rt.authTelegramWebApp)
		r.Group(func(r chi.Router) {
			r.Use(rt.jwtMiddleware)
			r.Get("/tasks/today", rt.listTasksToday)
			r.Get("/tasks", rt.listTasks)
			r.Post("/tasks", rt.createTask)
			r.Get("/tasks/{id}", rt.getTask)
			r.Patch("/tasks/{id}", rt.editTask)
			r.Delete("/tasks/{id}", rt.deleteTask)
			r.Post("/tasks/{id}/complete", rt.completeTask)
			r.Post("/tasks/{id}/cancel", rt.cancelTask)
			r.Post("/tasks/{id}/archive", rt.archiveTask)
			r.Post("/tasks/{id}/reschedule", rt.rescheduleTask)
			r.Get("/projects/progress", rt.projectProgress)
			r.Get("/reviews/morning", rt.morningReview)
			r.Get("/reviews/evening", rt.eveningReview)
			r.Get("/reviews/weekly", rt.weeklyReview)
			r.Get("/reviews/monthly", rt.monthlyReview)
			r.Get("/priorities", rt.listPriorities)
			r.Get("/analytics/summary", rt.analyticsSummary)
			r.Post("/finance/income", rt.recordIncome)
			r.Post("/finance/expense", rt.recordExpense)
			r.Get("/finance/cash-flow", rt.cashFlow)
			r.Get("/finance/debts", rt.listDebts)
			r.Post("/finance/debts", rt.createDebt)
			r.Post("/finance/debts/{id}/pay", rt.payDebt)
			r.Get("/habits/today", rt.listHabitsToday)
			r.Post("/habits", rt.createHabit)
			r.Post("/habits/{id}/track", rt.trackHabit)
			r.Post("/reminders", rt.createReminder)
			r.Get("/reminders", rt.listReminders)
			r.Delete("/reminders/{id}", rt.cancelReminder)
			r.Get("/notes", rt.listNotes)
			r.Post("/notes", rt.createNote)
			r.Delete("/notes/{id}", rt.deleteNote)
			r.Get("/career/contacts", rt.listContacts)
			r.Post("/career/contacts", rt.createContact)
			r.Delete("/career/contacts/{id}", rt.deleteContact)
			r.Get("/career/skills", rt.listSkills)
			r.Post("/career/skills", rt.createSkill)
			r.Delete("/career/skills/{id}", rt.deleteSkill)
			r.Get("/health/weight", rt.listWeights)
			r.Get("/health/weight/latest", rt.latestWeight)
			r.Post("/health/weight", rt.recordWeight)
			r.Get("/health/steps", rt.listSteps)
			r.Get("/health/steps/latest", rt.latestSteps)
			r.Post("/health/steps", rt.recordSteps)
			r.Get("/health/sleep", rt.listSleep)
			r.Get("/health/sleep/latest", rt.latestSleep)
			r.Post("/health/sleep", rt.recordSleep)
			r.Get("/calendar/today", rt.listCalendarToday)
			r.Post("/calendar/events", rt.createCalendarEvent)
			r.Get("/projects", rt.listProjects)
			r.Post("/projects", rt.createProject)
			r.Get("/projects/{id}/tasks", rt.listProjectTasks)
			r.Post("/projects/{id}/archive", rt.archiveProject)
			r.Get("/settings", rt.getSettings)
			r.Get("/settings/spheres", rt.listSpheres)
			r.Post("/settings/spheres", rt.createSphere)
			r.Put("/settings/spheres/{id}", rt.updateSphere)
			r.Delete("/settings/spheres/{id}", rt.deleteSphere)
			r.Put("/settings/morning-review", rt.updateMorningReview)
			r.Put("/settings/evening-review", rt.updateEveningReview)
			r.Put("/settings/quiet-hours", rt.updateQuietHours)
		})
	})
}

func (rt *Router) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rt.deps.Tokens == nil {
			writeError(w, http.StatusServiceUnavailable, "api auth not configured")
			return
		}
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		userID, err := rt.deps.Tokens.Parse(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
	})
}

type tokenRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (rt *Router) issueToken(w http.ResponseWriter, r *http.Request) {
	if rt.deps.Tokens == nil {
		writeError(w, http.StatusServiceUnavailable, "api auth not configured")
		return
	}
	if rt.deps.APIKey == "" {
		writeError(w, http.StatusServiceUnavailable, "api key not configured")
		return
	}
	key := r.Header.Get("X-API-Key")
	if key == "" || key != rt.deps.APIKey {
		writeError(w, http.StatusUnauthorized, "invalid api key")
		return
	}
	var req tokenRequest
	if err := decodeJSON(r, &req); err != nil || req.TelegramID <= 0 {
		writeError(w, http.StatusBadRequest, "telegram_id is required")
		return
	}
	user, err := rt.deps.GetUser.Execute(r.Context(), req.TelegramID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	token, exp, err := rt.deps.Tokens.Issue(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token issue failed")
		return
	}
	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: token,
		ExpiresIn:   int64(timeUntil(exp)),
		TokenType:   "Bearer",
	})
}

func timeUntil(exp time.Time) int {
	d := time.Until(exp)
	if d < 0 {
		return 0
	}
	return int(d.Seconds())
}

type taskJSON struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     *string  `json:"description,omitempty"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority"`
	DueDate         *string  `json:"due_date,omitempty"`
	DurationMinutes *int     `json:"duration_minutes,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	ProjectIDs      []string `json:"project_ids,omitempty"`
	CreatedAt       string   `json:"created_at"`
}

func taskToJSON(dto tasksapp.TaskDTO) taskJSON {
	out := taskJSON{
		ID:              dto.ID.String(),
		Title:           dto.Title,
		Description:     dto.Description,
		Status:          string(dto.Status),
		Priority:        string(dto.Priority),
		DurationMinutes: dto.DurationMinutes,
		Tags:            dto.Tags,
		ProjectIDs:      projectIDsToStrings(dto.ProjectIDs),
		CreatedAt:       dto.CreatedAt.UTC().Format(time.RFC3339),
	}
	if dto.DueDate != nil {
		s := dto.DueDate.Format("2006-01-02")
		out.DueDate = &s
	}
	return out
}

func projectIDsToStrings(ids []ids.ProjectID) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func (rt *Router) listTasksToday(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := rt.deps.ListToday.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]taskJSON, 0, len(items))
	for _, item := range items {
		out = append(out, taskToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": out})
}

type createTaskRequest struct {
	Title           string   `json:"title"`
	Priority        string   `json:"priority"`
	DueDate         *string  `json:"due_date"`
	DurationMinutes *int     `json:"duration_minutes"`
	Tags            []string `json:"tags"`
	ProjectIDs      []string `json:"project_ids"`
}

func (rt *Router) createTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createTaskRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	var due *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid due_date")
			return
		}
		due = &t
	}
	projectIDs, err := parseProjectIDs(req.ProjectIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project_ids")
		return
	}
	dto, err := rt.deps.CreateTask.Execute(r.Context(), tasksapp.CreateTaskInput{
		UserID:          userID,
		Title:           strings.TrimSpace(req.Title),
		Priority:        taskdomain.Priority(req.Priority),
		DueDate:         due,
		DurationMinutes: req.DurationMinutes,
		Tags:            req.Tags,
		ProjectIDs:      projectIDs,
		Source:          events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, taskToJSON(dto))
}

func (rt *Router) listTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	if tag == "" {
		writeError(w, http.StatusBadRequest, "tag query param is required")
		return
	}
	if rt.deps.ListByTag == nil {
		writeError(w, http.StatusNotImplemented, "list by tag is not configured")
		return
	}
	items, err := rt.deps.ListByTag.Execute(r.Context(), userID, tag)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out := make([]taskJSON, 0, len(items))
	for _, item := range items {
		out = append(out, taskToJSON(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": out})
}

type editTaskRequest struct {
	Title            *string   `json:"title"`
	Description      *string   `json:"description"`
	ClearDescription bool      `json:"clear_description"`
	Priority         *string   `json:"priority"`
	DueDate          *string   `json:"due_date"`
	ClearDueDate     bool      `json:"clear_due_date"`
	DurationMinutes  *int      `json:"duration_minutes"`
	ClearDuration    bool      `json:"clear_duration"`
	Tags             *[]string `json:"tags"`
	ProjectIDs       *[]string `json:"project_ids"`
}

func (rt *Router) editTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.EditTask == nil {
		writeError(w, http.StatusNotImplemented, "edit task is not configured")
		return
	}
	taskID, err := ids.ParseTaskID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var req editTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var due *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid due_date")
			return
		}
		due = &t
	}
	var priority *taskdomain.Priority
	if req.Priority != nil {
		p := taskdomain.Priority(*req.Priority)
		priority = &p
	}
	var projectIDs *[]ids.ProjectID
	if req.ProjectIDs != nil {
		parsed, err := parseProjectIDs(*req.ProjectIDs)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid project_ids")
			return
		}
		projectIDs = &parsed
	}
	dto, err := rt.deps.EditTask.Execute(r.Context(), tasksapp.EditTaskInput{
		UserID:           userID,
		TaskID:           taskID,
		Title:            req.Title,
		Description:      req.Description,
		ClearDescription: req.ClearDescription,
		Priority:         priority,
		DueDate:          due,
		ClearDueDate:     req.ClearDueDate,
		DurationMinutes:  req.DurationMinutes,
		ClearDuration:    req.ClearDuration,
		Tags:             req.Tags,
		ProjectIDs:       projectIDs,
		Source:           events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, taskToJSON(dto))
}

func (rt *Router) completeTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	taskID, err := ids.ParseTaskID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	dto, err := rt.deps.Complete.Execute(r.Context(), tasksapp.CompleteTaskInput{
		UserID: userID, TaskID: taskID, Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, taskToJSON(dto))
}

func (rt *Router) cancelTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.CancelTask == nil {
		writeError(w, http.StatusNotImplemented, "cancel task is not configured")
		return
	}
	taskID, err := ids.ParseTaskID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	dto, err := rt.deps.CancelTask.Execute(r.Context(), tasksapp.CancelTaskInput{
		UserID: userID, TaskID: taskID, Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, taskToJSON(dto))
}

type rescheduleTaskRequest struct {
	DueDate string `json:"due_date"`
}

func (rt *Router) rescheduleTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.RescheduleTask == nil {
		writeError(w, http.StatusNotImplemented, "reschedule task is not configured")
		return
	}
	taskID, err := ids.ParseTaskID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var req rescheduleTaskRequest
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.DueDate) == "" {
		writeError(w, http.StatusBadRequest, "due_date is required")
		return
	}
	due, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid due_date")
		return
	}
	dto, err := rt.deps.RescheduleTask.Execute(r.Context(), tasksapp.RescheduleTaskInput{
		UserID: userID, TaskID: taskID, DueDate: due, Source: events.SourceHTTP,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, taskToJSON(dto))
}

func (rt *Router) analyticsSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	summary, err := rt.deps.Analytics.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"period_label":      summary.PeriodLabel,
		"tasks_created":     summary.TasksCreated,
		"tasks_completed":   summary.TasksCompleted,
		"completion_rate":   summary.CompletionRate,
		"open_tasks":        summary.OpenTasks,
		"habit_consistency": summary.HabitConsistency,
		"habit_completions": summary.HabitCompletions,
		"habit_count":       summary.HabitCount,
		"projects":          summary.Projects,
	})
}

func parseProjectIDs(raw []string) ([]ids.ProjectID, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]ids.ProjectID, 0, len(raw))
	for _, s := range raw {
		id, err := ids.ParseProjectID(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}
