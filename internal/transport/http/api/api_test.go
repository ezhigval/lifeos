package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	financedomain "github.com/valentinezhov/lifeos/internal/finance/domain"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	habitsdomain "github.com/valentinezhov/lifeos/internal/habits/domain"
	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/identity/domain"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	knowledgedomain "github.com/valentinezhov/lifeos/internal/knowledge/domain"
	"github.com/valentinezhov/lifeos/internal/platform/auth"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	projectsdomain "github.com/valentinezhov/lifeos/internal/projects/domain"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
	"github.com/valentinezhov/lifeos/internal/transport/http/api"
)

const (
	testAPIKey    = "test-api-key"
	testJWTSecret = "test-secret-key-32bytes-min!!"
	testTelegram  = int64(900001)
)

type stubUserRepo struct {
	user domain.User
}

func (s *stubUserRepo) GetByTelegramID(_ context.Context, telegramID int64) (domain.User, error) {
	if telegramID != s.user.TelegramID {
		return domain.User{}, domain.ErrNotFound
	}
	return s.user, nil
}

func (s *stubUserRepo) Upsert(_ context.Context, user domain.User) error {
	s.user = user
	return nil
}

type fakeTaskStore struct {
	tasks map[ids.TaskID]taskdomain.Task
}

func newFakeTaskStore() *fakeTaskStore {
	return &fakeTaskStore{tasks: make(map[ids.TaskID]taskdomain.Task)}
}

func (s *fakeTaskStore) Save(_ context.Context, task taskdomain.Task) error {
	s.tasks[task.ID] = task
	return nil
}

func (s *fakeTaskStore) GetByID(_ context.Context, userID ids.UserID, taskID ids.TaskID) (taskdomain.Task, error) {
	task, ok := s.tasks[taskID]
	if !ok || task.UserID != userID {
		return taskdomain.Task{}, errors.New("not found")
	}
	return task, nil
}

func (s *fakeTaskStore) ListByDueDate(_ context.Context, userID ids.UserID, dueDate time.Time) ([]taskdomain.Task, error) {
	var out []taskdomain.Task
	for _, task := range s.tasks {
		if task.UserID != userID || task.DueDate == nil {
			continue
		}
		if task.DueDate.Format("2006-01-02") == dueDate.Format("2006-01-02") {
			out = append(out, task)
		}
	}
	return out, nil
}

func (s *fakeTaskStore) ListOpenDueOnOrBefore(_ context.Context, userID ids.UserID, dueDate time.Time) ([]taskdomain.Task, error) {
	var out []taskdomain.Task
	for _, task := range s.tasks {
		if task.UserID != userID || task.DueDate == nil {
			continue
		}
		if task.Status != taskdomain.StatusTodo && task.Status != taskdomain.StatusInProgress {
			continue
		}
		if !task.DueDate.After(dueDate) {
			out = append(out, task)
		}
	}
	return out, nil
}

func (s *fakeTaskStore) ListByTag(_ context.Context, userID ids.UserID, tag string) ([]taskdomain.Task, error) {
	var out []taskdomain.Task
	for _, task := range s.tasks {
		if task.UserID != userID {
			continue
		}
		for _, t := range task.Tags {
			if t == tag {
				out = append(out, task)
				break
			}
		}
	}
	return out, nil
}

func (s *fakeTaskStore) SetProjects(_ context.Context, taskID ids.TaskID, projectIDs []ids.ProjectID) error {
	task, ok := s.tasks[taskID]
	if !ok {
		return errors.New("not found")
	}
	task.ProjectIDs = projectIDs
	s.tasks[taskID] = task
	return nil
}

func (s *fakeTaskStore) ListByProject(context.Context, ids.UserID, ids.ProjectID) ([]taskdomain.Task, error) {
	return nil, nil
}

func (s *fakeTaskStore) Update(_ context.Context, task taskdomain.Task) error {
	s.tasks[task.ID] = task
	return nil
}

func (s *fakeTaskStore) FindOpenByTitle(context.Context, ids.UserID, string) (taskdomain.Task, error) {
	return taskdomain.Task{}, taskdomain.ErrNotFound
}

type fakeEvents struct{}

func (fakeEvents) Append(context.Context, events.Record) error { return nil }

type fakeTx struct{}

func (fakeTx) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeTZ struct{}

func (fakeTZ) Timezone(context.Context, ids.UserID) (string, error) {
	return "UTC", nil
}

type fakeProjectStore struct {
	projects map[ids.ProjectID]projectsdomain.Project
}

func newFakeProjectStore() *fakeProjectStore {
	return &fakeProjectStore{projects: make(map[ids.ProjectID]projectsdomain.Project)}
}

func (s *fakeProjectStore) Save(_ context.Context, project projectsdomain.Project) error {
	s.projects[project.ID] = project
	return nil
}

func (s *fakeProjectStore) SetSpheres(_ context.Context, _ ids.ProjectID, _ []ids.SphereID) error {
	return nil
}

func (s *fakeProjectStore) LoadSphereIDs(_ context.Context, projectID ids.ProjectID) ([]ids.SphereID, error) {
	p, ok := s.projects[projectID]
	if !ok {
		return nil, nil
	}
	return p.SphereIDs, nil
}

func (s *fakeProjectStore) FindByName(_ context.Context, userID ids.UserID, name string) (projectsdomain.Project, error) {
	for _, p := range s.projects {
		if p.UserID == userID && strings.EqualFold(p.Name, name) && p.Status == projectsdomain.StatusActive {
			return p, nil
		}
	}
	return projectsdomain.Project{}, projectsdomain.ErrNotFound
}

func (s *fakeProjectStore) GetByID(_ context.Context, userID ids.UserID, projectID ids.ProjectID) (projectsdomain.Project, error) {
	p, ok := s.projects[projectID]
	if !ok || p.UserID != userID {
		return projectsdomain.Project{}, projectsdomain.ErrNotFound
	}
	return p, nil
}

func (s *fakeProjectStore) ListActive(_ context.Context, userID ids.UserID) ([]projectsdomain.Project, error) {
	var out []projectsdomain.Project
	for _, p := range s.projects {
		if p.UserID == userID && p.Status == projectsdomain.StatusActive {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *fakeProjectStore) ListBySphere(_ context.Context, userID ids.UserID, sphereID ids.SphereID) ([]projectsdomain.Project, error) {
	var out []projectsdomain.Project
	for _, p := range s.projects {
		if p.UserID != userID || p.Status != projectsdomain.StatusActive {
			continue
		}
		for _, sid := range p.SphereIDs {
			if sid == sphereID {
				out = append(out, p)
				break
			}
		}
	}
	return out, nil
}

func (s *fakeProjectStore) Exists(_ context.Context, userID ids.UserID, projectID ids.ProjectID) (bool, error) {
	p, ok := s.projects[projectID]
	return ok && p.UserID == userID, nil
}

func (s *fakeProjectStore) AllExist(_ context.Context, userID ids.UserID, projectIDs []ids.ProjectID) (bool, error) {
	for _, id := range projectIDs {
		ok, err := s.Exists(context.Background(), userID, id)
		if err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

type fakeDebtStore struct {
	debts map[ids.DebtID]financedomain.Debt
}

func newFakeDebtStore() *fakeDebtStore {
	return &fakeDebtStore{debts: make(map[ids.DebtID]financedomain.Debt)}
}

func (s *fakeDebtStore) SaveDebt(_ context.Context, debt financedomain.Debt) error {
	s.debts[debt.ID] = debt
	return nil
}

func (s *fakeDebtStore) ListOpen(_ context.Context, userID ids.UserID) ([]financedomain.Debt, error) {
	var out []financedomain.Debt
	for _, d := range s.debts {
		if d.UserID == userID && d.Status == financedomain.DebtStatusOpen {
			out = append(out, d)
		}
	}
	return out, nil
}

type fakeNoteStore struct {
	notes []knowledgedomain.Note
}

func (s *fakeNoteStore) Save(_ context.Context, note knowledgedomain.Note) error {
	s.notes = append(s.notes, note)
	return nil
}

func (s *fakeNoteStore) ListRecent(_ context.Context, userID ids.UserID, limit int32) ([]knowledgedomain.Note, error) {
	var out []knowledgedomain.Note
	for i := len(s.notes) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		if s.notes[i].UserID == userID {
			out = append(out, s.notes[i])
		}
	}
	return out, nil
}

func (s *fakeNoteStore) ListByTag(_ context.Context, userID ids.UserID, tag string, limit int32) ([]knowledgedomain.Note, error) {
	tag = strings.ToLower(strings.TrimSpace(tag))
	var out []knowledgedomain.Note
	for i := len(s.notes) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		n := s.notes[i]
		if n.UserID != userID {
			continue
		}
		for _, t := range n.Tags {
			if t == tag {
				out = append(out, n)
				break
			}
		}
	}
	return out, nil
}

func (s *fakeNoteStore) Search(_ context.Context, userID ids.UserID, query string, limit int32) ([]knowledgedomain.Note, error) {
	needle := strings.ToLower(query)
	var out []knowledgedomain.Note
	for i := len(s.notes) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		n := s.notes[i]
		if n.UserID == userID && strings.Contains(strings.ToLower(n.Body), needle) {
			out = append(out, n)
		}
	}
	return out, nil
}

func (s *fakeNoteStore) Delete(_ context.Context, userID ids.UserID, noteID ids.NoteID) (knowledgedomain.Note, error) {
	for i, n := range s.notes {
		if n.ID == noteID && n.UserID == userID {
			s.notes = append(s.notes[:i], s.notes[i+1:]...)
			return n, nil
		}
	}
	return knowledgedomain.Note{}, knowledgedomain.ErrNotFound
}

type fakeHabitStore struct {
	habits map[ids.HabitID]habitsdomain.Habit
}

func newFakeHabitStore() *fakeHabitStore {
	return &fakeHabitStore{habits: make(map[ids.HabitID]habitsdomain.Habit)}
}

func (s *fakeHabitStore) Save(_ context.Context, habit habitsdomain.Habit) error {
	s.habits[habit.ID] = habit
	return nil
}

func (s *fakeHabitStore) GetByID(_ context.Context, userID ids.UserID, habitID ids.HabitID) (habitsdomain.Habit, error) {
	habit, ok := s.habits[habitID]
	if !ok || habit.UserID != userID {
		return habitsdomain.Habit{}, habitsdomain.ErrNotFound
	}
	return habit, nil
}

func (s *fakeHabitStore) FindByName(context.Context, ids.UserID, string) (habitsdomain.Habit, error) {
	return habitsdomain.Habit{}, habitsdomain.ErrNotFound
}

func (s *fakeHabitStore) ListWithToday(context.Context, ids.UserID, time.Time) ([]habitsapp.HabitDayRow, error) {
	return nil, nil
}

type testEnv struct {
	user   domain.User
	router chi.Router
}

func newTestEnv(t *testing.T) testEnv {
	t.Helper()
	user, err := domain.NewUser(testTelegram, "API User", "UTC", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.NewTokenService(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	store := newFakeTaskStore()
	create := tasksapp.NewCreateTask(store, fakeEvents{}, fakeTx{}, nil)
	complete := tasksapp.NewCompleteTask(store, fakeEvents{}, fakeTx{})
	listToday := tasksapp.NewListTasksToday(store, fakeTZ{})
	projectStore := newFakeProjectStore()
	habitStore := newFakeHabitStore()
	debtStore := newFakeDebtStore()
	noteStore := &fakeNoteStore{}
	users := &stubUserRepo{user: user}

	rt := api.NewRouter(api.Deps{
		Log:           slog.Default(),
		APIKey:        testAPIKey,
		BotToken:      "123456:TESTTOKEN",
		WebAppAuthTTL: time.Hour,
		Tokens:        tokens,
		GetUser:       identityapp.NewGetUserByTelegram(users),
		EnsureUser:    identityapp.NewEnsureUserByTelegram(users, nil, "UTC", nil),
		ListToday:     listToday,
		CreateTask:    create,
		Complete:      complete,
		CreateProject: projectsapp.NewCreateProject(projectStore, fakeEvents{}, fakeTx{}),
		ListProjects:  projectsapp.NewListProjects(projectStore),
		ProjectProg:   projectsapp.NewGetProjectProgress(projectStore),
		CreateHabit:   habitsapp.NewCreateHabit(habitStore, fakeEvents{}, fakeTx{}),
		CreateDebt:    financeapp.NewCreateDebt(debtStore, fakeEvents{}, fakeTx{}),
		ListDebts:     financeapp.NewListDebts(debtStore),
		CreateNote:    knowledgeapp.NewCreateNote(noteStore, fakeEvents{}, fakeTx{}),
		ListNotes:     knowledgeapp.NewListNotes(noteStore),
		SearchNotes:   knowledgeapp.NewSearchNotes(noteStore),
		DeleteNote:    knowledgeapp.NewDeleteNote(noteStore, fakeEvents{}, fakeTx{}),
	})
	r := chi.NewRouter()
	rt.Mount(r)
	return testEnv{user: user, router: r}
}

func doJSON(t *testing.T, handler http.Handler, method, path string, headers map[string]string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func issueToken(t *testing.T, env testEnv) string {
	t.Helper()
	rec := doJSON(t, env.router, http.MethodPost, "/api/v1/auth/token",
		map[string]string{"X-API-Key": testAPIKey},
		map[string]any{"telegram_id": testTelegram},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("issue token status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.AccessToken == "" {
		t.Fatal("empty access token")
	}
	return out.AccessToken
}

func TestIssueTokenRejectsMissingAPIKey(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	rec := doJSON(t, env.router, http.MethodPost, "/api/v1/auth/token", nil,
		map[string]any{"telegram_id": testTelegram},
	)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want 401", rec.Code)
	}
}

func TestAuthTelegramWebAppIssuesToken(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	const botToken = "123456:TESTTOKEN"
	now := time.Now().UTC()
	initData := auth.SignWebAppInitData(map[string]string{
		"auth_date": strconv.FormatInt(now.Unix(), 10),
		"user":      `{"id":900001,"first_name":"Test","username":"tester"}`,
	}, botToken)

	rec := doJSON(t, env.router, http.MethodPost, "/api/v1/auth/telegram-webapp", nil,
		map[string]any{"init_data": initData},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.AccessToken == "" || out.TokenType != "Bearer" {
		t.Fatalf("token response=%+v", out)
	}
}

func TestAuthTelegramWebAppRejectsBadInitData(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	rec := doJSON(t, env.router, http.MethodPost, "/api/v1/auth/telegram-webapp", nil,
		map[string]any{"init_data": "auth_date=1&hash=deadbeef"},
	)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want 401 body=%s", rec.Code, rec.Body.String())
	}
}

func TestIssueTokenRejectsUnknownUser(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	rec := doJSON(t, env.router, http.MethodPost, "/api/v1/auth/token",
		map[string]string{"X-API-Key": testAPIKey},
		map[string]any{"telegram_id": 1},
	)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
}

func TestProtectedRouteRequiresBearer(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	rec := doJSON(t, env.router, http.MethodGet, "/api/v1/tasks/today", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d want 401", rec.Code)
	}
}

func TestTaskHTTPOFlow(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	today := time.Now().UTC().Format("2006-01-02")
	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/tasks", auth, map[string]any{
		"title":    "api smoke task",
		"priority": "high",
		"due_date": today,
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	listRec := doJSON(t, env.router, http.MethodGet, "/api/v1/tasks/today", auth, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var listed struct {
		Tasks []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Tasks) != 1 || listed.Tasks[0].ID != created.ID {
		t.Fatalf("listed=%+v created=%s", listed.Tasks, created.ID)
	}

	completeRec := doJSON(t, env.router, http.MethodPost, "/api/v1/tasks/"+created.ID+"/complete", auth, nil)
	if completeRec.Code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", completeRec.Code, completeRec.Body.String())
	}
	var done struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(completeRec.Body.Bytes(), &done); err != nil {
		t.Fatal(err)
	}
	if done.Status != string(taskdomain.StatusDone) {
		t.Fatalf("status=%q want done", done.Status)
	}
}

func TestJWTRejectsWrongUserScope(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	other, err := auth.NewTokenService(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	foreignToken, _, err := other.Issue(ids.NewUserID())
	if err != nil {
		t.Fatal(err)
	}
	rec := doJSON(t, env.router, http.MethodGet, "/api/v1/tasks/today",
		map[string]string{"Authorization": "Bearer " + foreignToken}, nil,
	)
	if rec.Code != http.StatusOK {
		// foreign user simply sees empty list — isolation by user_id in use case
		t.Fatalf("status=%d", rec.Code)
	}
	var listed struct {
		Tasks []any `json:"tasks"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed.Tasks) != 0 {
		t.Fatalf("foreign user should see empty tasks, got %+v", listed.Tasks)
	}
	_ = env.user
}

func TestProjectHTTPFlow(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}
	sphereID := ids.NewSphereID()

	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/projects", auth, map[string]any{
		"name":         "накопить 500к",
		"sphere_ids":   []string{sphereID.String()},
		"target_value": "500000",
		"unit":         "RUB",
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	listRec := doJSON(t, env.router, http.MethodGet, "/api/v1/projects", auth, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var listed struct {
		Projects []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Projects) != 1 || listed.Projects[0].ID != created.ID {
		t.Fatalf("listed=%+v created=%s", listed.Projects, created.ID)
	}
}

func TestHabitHTTPFlow(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/habits", auth, map[string]any{
		"name": "бег",
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Name != "бег" {
		t.Fatalf("name=%q", created.Name)
	}
}

func TestCreateDebtHTTP(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/finance/debts", auth, map[string]any{
		"creditor":     "банку",
		"amount_cents": int64(1_000_000),
		"due_date":     "2026-12-31",
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		DueDate *string `json:"due_date"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.DueDate == nil || *created.DueDate != "2026-12-31" {
		t.Fatalf("due_date=%v", created.DueDate)
	}

	listRec := doJSON(t, env.router, http.MethodGet, "/api/v1/finance/debts", auth, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d", listRec.Code)
	}
	var listed struct {
		Debts []struct {
			Creditor string  `json:"creditor"`
			DueDate  *string `json:"due_date"`
		} `json:"debts"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Debts) != 1 || listed.Debts[0].Creditor != "банку" {
		t.Fatalf("listed=%+v", listed.Debts)
	}
	if listed.Debts[0].DueDate == nil || *listed.Debts[0].DueDate != "2026-12-31" {
		t.Fatalf("listed due_date=%v", listed.Debts[0].DueDate)
	}
}

func TestNoteHTTPFlow(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/notes", auth, map[string]any{
		"body": "#work идея для Jarvis",
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID   string   `json:"id"`
		Body string   `json:"body"`
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Body != "идея для Jarvis" || len(created.Tags) != 1 || created.Tags[0] != "work" {
		t.Fatalf("created=%+v", created)
	}

	listRec := doJSON(t, env.router, http.MethodGet, "/api/v1/notes", auth, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d", listRec.Code)
	}
	var listed struct {
		Notes []struct {
			Body string `json:"body"`
		} `json:"notes"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Notes) != 1 || listed.Notes[0].Body != "идея для Jarvis" {
		t.Fatalf("listed=%+v", listed.Notes)
	}

	tagRec := doJSON(t, env.router, http.MethodGet, "/api/v1/notes?tag=work", auth, nil)
	if tagRec.Code != http.StatusOK {
		t.Fatalf("tag list status=%d", tagRec.Code)
	}
	var tagged struct {
		Notes []struct {
			Body string `json:"body"`
		} `json:"notes"`
		Tag string `json:"tag"`
	}
	if err := json.Unmarshal(tagRec.Body.Bytes(), &tagged); err != nil {
		t.Fatal(err)
	}
	if tagged.Tag != "work" || len(tagged.Notes) != 1 {
		t.Fatalf("tagged=%+v", tagged)
	}

	searchRec := doJSON(t, env.router, http.MethodGet, "/api/v1/notes?q=jarvis", auth, nil)
	if searchRec.Code != http.StatusOK {
		t.Fatalf("search status=%d", searchRec.Code)
	}
	var searched struct {
		Notes []struct {
			Body string `json:"body"`
		} `json:"notes"`
		Query string `json:"query"`
	}
	if err := json.Unmarshal(searchRec.Body.Bytes(), &searched); err != nil {
		t.Fatal(err)
	}
	if searched.Query != "jarvis" || len(searched.Notes) != 1 {
		t.Fatalf("searched=%+v", searched)
	}

	delRec := doJSON(t, env.router, http.MethodDelete, "/api/v1/notes/"+created.ID, auth, nil)
	if delRec.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", delRec.Code, delRec.Body.String())
	}
	listAfter := doJSON(t, env.router, http.MethodGet, "/api/v1/notes", auth, nil)
	var after struct {
		Notes []any `json:"notes"`
	}
	if err := json.Unmarshal(listAfter.Body.Bytes(), &after); err != nil {
		t.Fatal(err)
	}
	if len(after.Notes) != 0 {
		t.Fatalf("notes after delete=%d", len(after.Notes))
	}
}
