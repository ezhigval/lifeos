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
	"github.com/google/uuid"

	calendarapp "github.com/valentinezhov/lifeos/internal/calendar/app"
	calendardomain "github.com/valentinezhov/lifeos/internal/calendar/domain"
	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	careerdomain "github.com/valentinezhov/lifeos/internal/career/domain"
	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	financedomain "github.com/valentinezhov/lifeos/internal/finance/domain"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	habitsdomain "github.com/valentinezhov/lifeos/internal/habits/domain"
	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	healthdomain "github.com/valentinezhov/lifeos/internal/health/domain"
	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/identity/domain"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	knowledgedomain "github.com/valentinezhov/lifeos/internal/knowledge/domain"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/auth"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	projectsdomain "github.com/valentinezhov/lifeos/internal/projects/domain"
	"github.com/valentinezhov/lifeos/internal/query"
	settingsapp "github.com/valentinezhov/lifeos/internal/settings/app"
	settingsdomain "github.com/valentinezhov/lifeos/internal/settings/domain"
	settingsinfra "github.com/valentinezhov/lifeos/internal/settings/infra"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	spheresdomain "github.com/valentinezhov/lifeos/internal/spheres/domain"
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

func (s *stubUserRepo) GetByID(_ context.Context, userID ids.UserID) (domain.User, error) {
	if userID != s.user.ID {
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
	if !ok || task.UserID != userID || task.DeletedAt != nil {
		return taskdomain.Task{}, taskdomain.ErrNotFound
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

func (s *fakeTaskStore) ListOpenDueBetween(context.Context, ids.UserID, time.Time, time.Time) ([]taskdomain.Task, error) {
	return nil, nil
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
	plans map[ids.PlannedCashflowID]financedomain.PlannedCashflow
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

func (s *fakeDebtStore) GetByID(_ context.Context, userID ids.UserID, debtID ids.DebtID) (financedomain.Debt, error) {
	d, ok := s.debts[debtID]
	if !ok || d.UserID != userID {
		return financedomain.Debt{}, financedomain.ErrDebtNotFound
	}
	return d, nil
}

func (s *fakeDebtStore) FindOpenByCreditor(_ context.Context, userID ids.UserID, creditor string) (financedomain.Debt, error) {
	for _, d := range s.debts {
		if d.UserID == userID && d.Status == financedomain.DebtStatusOpen && d.Creditor == creditor {
			return d, nil
		}
	}
	return financedomain.Debt{}, financedomain.ErrDebtNotFound
}

func (s *fakeDebtStore) UpdateDebt(_ context.Context, debt financedomain.Debt) error {
	s.debts[debt.ID] = debt
	return nil
}

func (s *fakeDebtStore) SavePlanned(_ context.Context, item financedomain.PlannedCashflow) error {
	if s.plans == nil {
		s.plans = make(map[ids.PlannedCashflowID]financedomain.PlannedCashflow)
	}
	s.plans[item.ID] = item
	return nil
}

func (s *fakeDebtStore) GetPlanned(_ context.Context, userID ids.UserID, id ids.PlannedCashflowID) (financedomain.PlannedCashflow, error) {
	p, ok := s.plans[id]
	if !ok || p.UserID != userID {
		return financedomain.PlannedCashflow{}, financedomain.ErrPlanNotFound
	}
	return p, nil
}

func (s *fakeDebtStore) UpdatePlannedNextDate(_ context.Context, item financedomain.PlannedCashflow) error {
	p, ok := s.plans[item.ID]
	if !ok || p.UserID != item.UserID {
		return financedomain.ErrPlanNotFound
	}
	p.NextDate = item.NextDate
	p.UpdatedAt = item.UpdatedAt
	s.plans[item.ID] = p
	return nil
}

func (s *fakeDebtStore) ListPlanned(_ context.Context, userID ids.UserID) ([]financedomain.PlannedCashflow, error) {
	var out []financedomain.PlannedCashflow
	for _, p := range s.plans {
		if p.UserID == userID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *fakeDebtStore) DeletePlanned(_ context.Context, userID ids.UserID, id ids.PlannedCashflowID) error {
	p, ok := s.plans[id]
	if !ok || p.UserID != userID {
		return financedomain.ErrPlanNotFound
	}
	delete(s.plans, id)
	return nil
}

type fakeNoteStore struct {
	notes []knowledgedomain.Note
}

func (s *fakeNoteStore) Save(_ context.Context, note knowledgedomain.Note) error {
	s.notes = append(s.notes, note)
	return nil
}

func (s *fakeNoteStore) GetByID(_ context.Context, userID ids.UserID, noteID ids.NoteID) (knowledgedomain.Note, error) {
	for _, n := range s.notes {
		if n.ID == noteID && n.UserID == userID {
			return n, nil
		}
	}
	return knowledgedomain.Note{}, knowledgedomain.ErrNotFound
}

func (s *fakeNoteStore) UpdateBody(_ context.Context, userID ids.UserID, noteID ids.NoteID, body string, now time.Time) (knowledgedomain.Note, error) {
	for i, n := range s.notes {
		if n.ID == noteID && n.UserID == userID {
			n.Body = body
			n.UpdatedAt = now
			s.notes[i] = n
			return n, nil
		}
	}
	return knowledgedomain.Note{}, knowledgedomain.ErrNotFound
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

type fakeContactStore struct {
	contacts []careerdomain.Contact
}

func (s *fakeContactStore) Save(_ context.Context, contact careerdomain.Contact) error {
	s.contacts = append(s.contacts, contact)
	return nil
}

func (s *fakeContactStore) ListRecent(_ context.Context, userID ids.UserID, limit int32) ([]careerdomain.Contact, error) {
	var out []careerdomain.Contact
	for i := len(s.contacts) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		if s.contacts[i].UserID == userID {
			out = append(out, s.contacts[i])
		}
	}
	return out, nil
}

func (s *fakeContactStore) Search(_ context.Context, userID ids.UserID, query string, limit int32) ([]careerdomain.Contact, error) {
	needle := strings.ToLower(query)
	var out []careerdomain.Contact
	for i := len(s.contacts) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		c := s.contacts[i]
		if c.UserID != userID {
			continue
		}
		hay := strings.ToLower(c.Name + " " + c.Company + " " + c.Role)
		if strings.Contains(hay, needle) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (s *fakeContactStore) Delete(_ context.Context, userID ids.UserID, contactID ids.ContactID) (careerdomain.Contact, error) {
	for i, c := range s.contacts {
		if c.ID == contactID && c.UserID == userID {
			s.contacts = append(s.contacts[:i], s.contacts[i+1:]...)
			return c, nil
		}
	}
	return careerdomain.Contact{}, careerdomain.ErrNotFound
}

type fakeSkillStore struct {
	skills []careerdomain.Skill
}

func (s *fakeSkillStore) SaveSkill(_ context.Context, skill careerdomain.Skill) error {
	s.skills = append(s.skills, skill)
	return nil
}

func (s *fakeSkillStore) ListRecentSkills(_ context.Context, userID ids.UserID, limit int32) ([]careerdomain.Skill, error) {
	var out []careerdomain.Skill
	for i := len(s.skills) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		if s.skills[i].UserID == userID {
			out = append(out, s.skills[i])
		}
	}
	return out, nil
}

func (s *fakeSkillStore) SearchSkills(_ context.Context, userID ids.UserID, query string, limit int32) ([]careerdomain.Skill, error) {
	needle := strings.ToLower(query)
	var out []careerdomain.Skill
	for i := len(s.skills) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		sk := s.skills[i]
		if sk.UserID == userID && strings.Contains(strings.ToLower(sk.Name), needle) {
			out = append(out, sk)
		}
	}
	return out, nil
}

func (s *fakeSkillStore) DeleteSkill(_ context.Context, userID ids.UserID, skillID ids.SkillID) (careerdomain.Skill, error) {
	for i, sk := range s.skills {
		if sk.ID == skillID && sk.UserID == userID {
			s.skills = append(s.skills[:i], s.skills[i+1:]...)
			return sk, nil
		}
	}
	return careerdomain.Skill{}, careerdomain.ErrSkillNotFound
}

type fakeHealthStore struct {
	weights []healthdomain.WeightLog
	steps   []healthdomain.StepLog
	sleep   []healthdomain.SleepLog
}

func (s *fakeHealthStore) Save(_ context.Context, log healthdomain.WeightLog) error {
	s.weights = append(s.weights, log)
	return nil
}

func (s *fakeHealthStore) GetLatest(_ context.Context, userID ids.UserID) (healthdomain.WeightLog, error) {
	for i := len(s.weights) - 1; i >= 0; i-- {
		if s.weights[i].UserID == userID {
			return s.weights[i], nil
		}
	}
	return healthdomain.WeightLog{}, healthdomain.ErrNotFound
}

func (s *fakeHealthStore) ListRecent(_ context.Context, userID ids.UserID, limit int32) ([]healthdomain.WeightLog, error) {
	var out []healthdomain.WeightLog
	for i := len(s.weights) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		if s.weights[i].UserID == userID {
			out = append(out, s.weights[i])
		}
	}
	return out, nil
}

func (s *fakeHealthStore) SaveSteps(_ context.Context, log healthdomain.StepLog) error {
	s.steps = append(s.steps, log)
	return nil
}

func (s *fakeHealthStore) GetLatestSteps(_ context.Context, userID ids.UserID) (healthdomain.StepLog, error) {
	for i := len(s.steps) - 1; i >= 0; i-- {
		if s.steps[i].UserID == userID {
			return s.steps[i], nil
		}
	}
	return healthdomain.StepLog{}, healthdomain.ErrNotFound
}

func (s *fakeHealthStore) ListRecentSteps(_ context.Context, userID ids.UserID, limit int32) ([]healthdomain.StepLog, error) {
	var out []healthdomain.StepLog
	for i := len(s.steps) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		if s.steps[i].UserID == userID {
			out = append(out, s.steps[i])
		}
	}
	return out, nil
}

func (s *fakeHealthStore) SaveSleep(_ context.Context, log healthdomain.SleepLog) error {
	s.sleep = append(s.sleep, log)
	return nil
}

func (s *fakeHealthStore) GetLatestSleep(_ context.Context, userID ids.UserID) (healthdomain.SleepLog, error) {
	for i := len(s.sleep) - 1; i >= 0; i-- {
		if s.sleep[i].UserID == userID {
			return s.sleep[i], nil
		}
	}
	return healthdomain.SleepLog{}, healthdomain.ErrNotFound
}

func (s *fakeHealthStore) ListRecentSleep(_ context.Context, userID ids.UserID, limit int32) ([]healthdomain.SleepLog, error) {
	var out []healthdomain.SleepLog
	for i := len(s.sleep) - 1; i >= 0 && int32(len(out)) < limit; i-- {
		if s.sleep[i].UserID == userID {
			out = append(out, s.sleep[i])
		}
	}
	return out, nil
}

type fakeReminderSvc struct {
	items []notifapp.ReminderDTO
}

func (s *fakeReminderSvc) ExecuteSchedule(_ context.Context, in notifapp.ScheduleReminderInput) (notifapp.ReminderDTO, error) {
	dto := notifapp.ReminderDTO{
		ID:      uuid.Must(uuid.NewV7()).String(),
		Message: in.Message,
		FireAt:  in.FireAt,
		Status:  "pending",
	}
	s.items = append(s.items, dto)
	return dto, nil
}

type scheduleReminderAdapter struct{ *fakeReminderSvc }

func (a scheduleReminderAdapter) Execute(ctx context.Context, in notifapp.ScheduleReminderInput) (notifapp.ReminderDTO, error) {
	return a.ExecuteSchedule(ctx, in)
}

func (s *fakeReminderSvc) ExecuteList(_ context.Context, _ ids.UserID) ([]notifapp.ReminderDTO, error) {
	out := make([]notifapp.ReminderDTO, 0, len(s.items))
	for _, item := range s.items {
		if item.Status == "pending" {
			out = append(out, item)
		}
	}
	return out, nil
}

type listReminderAdapter struct{ *fakeReminderSvc }

func (a listReminderAdapter) Execute(ctx context.Context, userID ids.UserID) ([]notifapp.ReminderDTO, error) {
	return a.ExecuteList(ctx, userID)
}

func (s *fakeReminderSvc) ExecuteCancel(_ context.Context, in notifapp.CancelReminderInput) (notifapp.ReminderDTO, error) {
	id := in.ReminderID.String()
	for i, item := range s.items {
		if item.ID == id && item.Status == "pending" {
			item.Status = "cancelled"
			s.items[i] = item
			return item, nil
		}
	}
	return notifapp.ReminderDTO{}, notifapp.ErrReminderNotFound
}

type cancelReminderAdapter struct{ *fakeReminderSvc }

func (a cancelReminderAdapter) Execute(ctx context.Context, in notifapp.CancelReminderInput) (notifapp.ReminderDTO, error) {
	return a.ExecuteCancel(ctx, in)
}

func (a cancelReminderAdapter) CancelForTask(_ context.Context, _ ids.UserID, _ string) error {
	return nil
}

type fakeAnalytics struct {
	summary query.ProductivitySummary
}

func (f fakeAnalytics) Execute(context.Context, ids.UserID) (query.ProductivitySummary, error) {
	return f.summary, nil
}

type fakeHabitStore struct {
	habits map[ids.HabitID]habitsdomain.Habit
	logs   map[ids.HabitID][]habitsdomain.HabitLog
}

func newFakeHabitStore() *fakeHabitStore {
	return &fakeHabitStore{
		habits: make(map[ids.HabitID]habitsdomain.Habit),
		logs:   make(map[ids.HabitID][]habitsdomain.HabitLog),
	}
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

func (s *fakeHabitStore) ListWithToday(_ context.Context, userID ids.UserID, today time.Time) ([]habitsapp.HabitDayRow, error) {
	day := today.Format("2006-01-02")
	out := make([]habitsapp.HabitDayRow, 0, len(s.habits))
	for _, habit := range s.habits {
		if habit.UserID != userID {
			continue
		}
		completed := false
		for _, log := range s.logs[habit.ID] {
			if log.Completed && log.LogDate.Format("2006-01-02") == day {
				completed = true
				break
			}
		}
		out = append(out, habitsapp.HabitDayRow{Habit: habit, TodayCompleted: completed})
	}
	return out, nil
}

func (s *fakeHabitStore) Upsert(_ context.Context, log habitsdomain.HabitLog) error {
	logs := s.logs[log.HabitID]
	for i, existing := range logs {
		if existing.LogDate.Format("2006-01-02") == log.LogDate.Format("2006-01-02") {
			logs[i] = log
			s.logs[log.HabitID] = logs
			return nil
		}
	}
	s.logs[log.HabitID] = append(logs, log)
	return nil
}

func (s *fakeHabitStore) ListSince(_ context.Context, habitID ids.HabitID, since time.Time) ([]habitsdomain.HabitLog, error) {
	var out []habitsdomain.HabitLog
	for _, log := range s.logs[habitID] {
		if !log.LogDate.Before(since) {
			out = append(out, log)
		}
	}
	return out, nil
}

type fakeCalendarStore struct {
	events []calendardomain.Event
}

func (s *fakeCalendarStore) Save(_ context.Context, event calendardomain.Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *fakeCalendarStore) ListBetween(_ context.Context, userID ids.UserID, from, to time.Time) ([]calendardomain.Event, error) {
	var out []calendardomain.Event
	for _, e := range s.events {
		if e.UserID != userID {
			continue
		}
		if !e.StartsAt.Before(from) && e.StartsAt.Before(to) {
			out = append(out, e)
		}
	}
	return out, nil
}

type fakeSettingsStore struct {
	byUser map[ids.UserID]settingsdomain.UserSettings
}

func newFakeSettingsStore() *fakeSettingsStore {
	return &fakeSettingsStore{byUser: make(map[ids.UserID]settingsdomain.UserSettings)}
}

func (s *fakeSettingsStore) Get(_ context.Context, userID ids.UserID) (settingsdomain.UserSettings, error) {
	if settings, ok := s.byUser[userID]; ok {
		return settings, nil
	}
	settings := settingsdomain.DefaultSettings(userID)
	s.byUser[userID] = settings
	return settings, nil
}

func (s *fakeSettingsStore) UpdateMorningReview(_ context.Context, userID ids.UserID, at settingsdomain.TimeOfDay) error {
	settings, _ := s.Get(context.Background(), userID)
	settings.MorningReviewAt = at
	s.byUser[userID] = settings
	return nil
}

func (s *fakeSettingsStore) UpdateEveningReview(_ context.Context, userID ids.UserID, at settingsdomain.TimeOfDay) error {
	settings, _ := s.Get(context.Background(), userID)
	settings.EveningReviewAt = at
	s.byUser[userID] = settings
	return nil
}

func (s *fakeSettingsStore) UpdateQuietHours(_ context.Context, userID ids.UserID, start, end settingsdomain.TimeOfDay) error {
	settings, _ := s.Get(context.Background(), userID)
	settings.QuietHoursStart = &start
	settings.QuietHoursEnd = &end
	s.byUser[userID] = settings
	return nil
}

type noopReviewRescheduler struct{}

func (noopReviewRescheduler) RescheduleReview(context.Context, ids.UserID, string, time.Time) error {
	return nil
}

type fakeSphereStore struct {
	spheres map[ids.SphereID]spheresdomain.Sphere
}

func newFakeSphereStore() *fakeSphereStore {
	return &fakeSphereStore{spheres: make(map[ids.SphereID]spheresdomain.Sphere)}
}

func (s *fakeSphereStore) Save(_ context.Context, sphere spheresdomain.Sphere) error {
	s.spheres[sphere.ID] = sphere
	return nil
}

func (s *fakeSphereStore) List(_ context.Context, userID ids.UserID) ([]spheresdomain.Sphere, error) {
	out := make([]spheresdomain.Sphere, 0, len(s.spheres))
	for _, sphere := range s.spheres {
		if sphere.UserID == userID {
			out = append(out, sphere)
		}
	}
	return out, nil
}

func (s *fakeSphereStore) Get(_ context.Context, userID ids.UserID, sphereID ids.SphereID) (spheresdomain.Sphere, error) {
	sphere, ok := s.spheres[sphereID]
	if !ok || sphere.UserID != userID {
		return spheresdomain.Sphere{}, spheresdomain.ErrNotFound
	}
	return sphere, nil
}

func (s *fakeSphereStore) FindByName(_ context.Context, userID ids.UserID, name string) (spheresdomain.Sphere, error) {
	for _, sphere := range s.spheres {
		if sphere.UserID == userID && sphere.Name == name {
			return sphere, nil
		}
	}
	return spheresdomain.Sphere{}, spheresdomain.ErrNotFound
}

func (s *fakeSphereStore) Count(_ context.Context, userID ids.UserID) (int32, error) {
	var n int32
	for _, sphere := range s.spheres {
		if sphere.UserID == userID {
			n++
		}
	}
	return n, nil
}

func (s *fakeSphereStore) Update(_ context.Context, sphere spheresdomain.Sphere) error {
	s.spheres[sphere.ID] = sphere
	return nil
}

func (s *fakeSphereStore) Delete(_ context.Context, userID ids.UserID, sphereID ids.SphereID) (spheresdomain.Sphere, error) {
	sphere, ok := s.spheres[sphereID]
	if !ok || sphere.UserID != userID {
		return spheresdomain.Sphere{}, spheresdomain.ErrNotFound
	}
	delete(s.spheres, sphereID)
	return sphere, nil
}

func (s *fakeSphereStore) HasLinkedProjects(context.Context, ids.SphereID) (bool, error) {
	return false, nil
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
	cancel := tasksapp.NewCancelTask(store, fakeEvents{}, fakeTx{})
	edit := tasksapp.NewEditTask(store, fakeEvents{}, fakeTx{}, nil)
	reschedule := tasksapp.NewRescheduleTask(store, fakeEvents{}, fakeTx{})
	listToday := tasksapp.NewListTasksToday(store, fakeTZ{})
	listByTag := tasksapp.NewListTasksByTag(store)
	getTask := tasksapp.NewGetTask(store)
	archive := tasksapp.NewArchiveTask(store, fakeEvents{}, fakeTx{})
	deleteTask := tasksapp.NewDeleteTask(store, fakeEvents{}, fakeTx{})
	projectStore := newFakeProjectStore()
	habitStore := newFakeHabitStore()
	calendarStore := &fakeCalendarStore{}
	settingsStore := newFakeSettingsStore()
	sphereStore := newFakeSphereStore()
	debtStore := newFakeDebtStore()
	noteStore := &fakeNoteStore{}
	contactStore := &fakeContactStore{}
	skillStore := &fakeSkillStore{}
	healthStore := &fakeHealthStore{}
	reminderSvc := &fakeReminderSvc{}
	users := &stubUserRepo{user: user}
	tzFn := func(context.Context, ids.UserID) (string, error) { return "UTC", nil }

	rt := api.NewRouter(api.Deps{
		Log:            slog.Default(),
		APIKey:         testAPIKey,
		BotToken:       "123456:TESTTOKEN",
		WebAppAuthTTL:  time.Hour,
		Tokens:         tokens,
		GetUser:        identityapp.NewGetUserByTelegram(users),
		GetUserByID:    identityapp.NewGetUserByID(users),
		EnsureUser:     identityapp.NewEnsureUserByTelegram(users, nil, "UTC", nil),
		ListToday:      listToday,
		CreateTask:     create,
		Complete:       complete,
		CancelTask:     cancel,
		EditTask:       edit,
		RescheduleTask: reschedule,
		ListByTag:      listByTag,
		GetTask:        getTask,
		ArchiveTask:    archive,
		DeleteTask:     deleteTask,
		CreateProject:  projectsapp.NewCreateProject(projectStore, fakeEvents{}, fakeTx{}),
		ListProjects:   projectsapp.NewListProjects(projectStore),
		ProjectProg:    projectsapp.NewGetProjectProgress(projectStore),
		ListHabits:     habitsapp.NewListHabitsToday(habitStore, habitStore, fakeTZ{}),
		CreateHabit:    habitsapp.NewCreateHabit(habitStore, fakeEvents{}, fakeTx{}),
		TrackHabit:     habitsapp.NewTrackHabit(habitStore, habitStore, fakeEvents{}, fakeTx{}, fakeTZ{}),
		ListCalendar:   calendarapp.NewListEventsToday(calendarStore, fakeTZ{}),
		CreateEvent:    calendarapp.NewCreateEvent(calendarStore, fakeEvents{}, fakeTx{}),
		GetSettings:    settingsapp.NewGetSettings(settingsStore),
		UpdateMorning: settingsapp.NewUpdateMorningReview(
			settingsStore, noopReviewRescheduler{}, tzFn, settingsinfra.ReviewAt,
		),
		UpdateEvening: settingsapp.NewUpdateEveningReview(
			settingsStore, noopReviewRescheduler{}, tzFn, settingsinfra.ReviewAt,
		),
		UpdateQuiet:      settingsapp.NewUpdateQuietHours(settingsStore),
		CreateSphere:     spheresapp.NewCreateSphere(sphereStore, fakeEvents{}, fakeTx{}),
		ListSpheres:      spheresapp.NewListSpheres(sphereStore),
		UpdateSphere:     spheresapp.NewUpdateSphere(sphereStore, fakeEvents{}, fakeTx{}),
		DeleteSphere:     spheresapp.NewDeleteSphere(sphereStore, fakeEvents{}, fakeTx{}),
		CreateDebt:       financeapp.NewCreateDebt(debtStore, fakeEvents{}, fakeTx{}),
		ListDebts:        financeapp.NewListDebts(debtStore),
		PayDebt:          financeapp.NewPayDebt(debtStore, fakeEvents{}, fakeTx{}),
		ListFinancePlan:  financeapp.NewListFinancePlan(debtStore, debtStore),
		CreatePlanned:    financeapp.NewCreatePlannedCashflow(debtStore, fakeEvents{}, fakeTx{}),
		DeletePlanned:    financeapp.NewDeletePlannedCashflow(debtStore),
		CompletePlanned:  financeapp.NewCompletePlanOccurrence(debtStore, fakeEvents{}, fakeTx{}),
		CreateNote:       knowledgeapp.NewCreateNote(noteStore, fakeEvents{}, fakeTx{}),
		ListNotes:        knowledgeapp.NewListNotes(noteStore),
		SearchNotes:      knowledgeapp.NewSearchNotes(noteStore),
		GetNote:          knowledgeapp.NewGetNote(noteStore),
		UpdateNote:       knowledgeapp.NewUpdateNote(noteStore),
		DeleteNote:       knowledgeapp.NewDeleteNote(noteStore, fakeEvents{}, fakeTx{}),
		CreateContact:    careerapp.NewCreateContact(contactStore, fakeEvents{}, fakeTx{}),
		ListContacts:     careerapp.NewListContacts(contactStore),
		SearchContacts:   careerapp.NewSearchContacts(contactStore),
		DeleteContact:    careerapp.NewDeleteContact(contactStore, fakeEvents{}, fakeTx{}),
		CreateSkill:      careerapp.NewCreateSkill(skillStore, fakeEvents{}, fakeTx{}),
		ListSkills:       careerapp.NewListSkills(skillStore),
		SearchSkills:     careerapp.NewSearchSkills(skillStore),
		DeleteSkill:      careerapp.NewDeleteSkill(skillStore, fakeEvents{}, fakeTx{}),
		RecordWeight:     healthapp.NewRecordWeight(healthStore, fakeEvents{}, fakeTx{}),
		GetLatestWeight:  healthapp.NewGetLatestWeight(healthStore),
		ListWeights:      healthapp.NewListWeights(healthStore),
		RecordSteps:      healthapp.NewRecordSteps(healthStore, fakeEvents{}, fakeTx{}),
		GetLatestSteps:   healthapp.NewGetLatestSteps(healthStore),
		ListSteps:        healthapp.NewListSteps(healthStore),
		RecordSleep:      healthapp.NewRecordSleep(healthStore, fakeEvents{}, fakeTx{}),
		GetLatestSleep:   healthapp.NewGetLatestSleep(healthStore),
		ListSleep:        healthapp.NewListSleep(healthStore),
		ScheduleReminder: scheduleReminderAdapter{reminderSvc},
		ListReminders:    listReminderAdapter{reminderSvc},
		CancelReminder:   cancelReminderAdapter{reminderSvc},
		Analytics: fakeAnalytics{summary: query.ProductivitySummary{
			PeriodLabel:      "июль 2026",
			TasksCreated:     10,
			TasksCompleted:   7,
			CompletionRate:   70,
			OpenTasks:        3,
			HabitConsistency: 80,
			HabitCompletions: 16,
			HabitCount:       2,
			Projects: []query.ProjectKPI{
				{Title: "LifeOS", Percent: "42%"},
			},
		}},
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
	const telegramID int64 = 900001
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
		TelegramID  int64  `json:"telegram_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.AccessToken == "" || out.TokenType != "Bearer" {
		t.Fatalf("token response=%+v", out)
	}
	if out.TelegramID != telegramID {
		t.Fatalf("telegram_id=%d want %d (must come from signed initData.user.id)", out.TelegramID, telegramID)
	}

	// Protected route must accept the issued JWT (LifeOS user_id derived from that telegram id).
	authHdr := map[string]string{"Authorization": "Bearer " + out.AccessToken}
	today := doJSON(t, env.router, http.MethodGet, "/api/v1/tasks/today", authHdr, nil)
	if today.Code != http.StatusOK {
		t.Fatalf("tasks/today status=%d body=%s", today.Code, today.Body.String())
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

func TestTaskLifecycleEditClearArchiveDelete(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	today := time.Now().UTC().Format("2006-01-02")
	desc := "body"
	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/tasks", auth, map[string]any{
		"title":       "edit me #work",
		"description": desc,
		"priority":    "medium",
		"due_date":    today,
		"tags":        []string{"inbox"},
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID          string  `json:"id"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
		Tags        []string `json:"tags"`
		DueDate     *string `json:"due_date"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Title != "edit me" || created.Description == nil || *created.Description != desc {
		t.Fatalf("created=%+v", created)
	}
	tagFound := false
	for _, tag := range created.Tags {
		if tag == "work" || tag == "inbox" {
			tagFound = true
		}
	}
	if !tagFound {
		t.Fatalf("tags=%v", created.Tags)
	}

	getRec := doJSON(t, env.router, http.MethodGet, "/api/v1/tasks/"+created.ID, auth, nil)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", getRec.Code, getRec.Body.String())
	}

	// Mini App style PATCH: explicit null clears due_date and description.
	patchRec := doJSON(t, env.router, http.MethodPatch, "/api/v1/tasks/"+created.ID, auth, map[string]any{
		"title":       "edited",
		"priority":    "high",
		"due_date":    nil,
		"description": nil,
	})
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", patchRec.Code, patchRec.Body.String())
	}
	var patched struct {
		Title       string  `json:"title"`
		Priority    string  `json:"priority"`
		Description *string `json:"description"`
		DueDate     *string `json:"due_date"`
	}
	if err := json.Unmarshal(patchRec.Body.Bytes(), &patched); err != nil {
		t.Fatal(err)
	}
	if patched.Title != "edited" || patched.Priority != "high" {
		t.Fatalf("patched=%+v", patched)
	}
	if patched.Description != nil || patched.DueDate != nil {
		t.Fatalf("expected cleared description/due_date, got %+v", patched)
	}

	listByTag := doJSON(t, env.router, http.MethodGet, "/api/v1/tasks?tag=work", auth, nil)
	if listByTag.Code != http.StatusOK {
		t.Fatalf("list by tag status=%d body=%s", listByTag.Code, listByTag.Body.String())
	}

	archiveRec := doJSON(t, env.router, http.MethodPost, "/api/v1/tasks/"+created.ID+"/archive", auth, nil)
	if archiveRec.Code != http.StatusOK {
		t.Fatalf("archive status=%d body=%s", archiveRec.Code, archiveRec.Body.String())
	}
	var archived struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(archiveRec.Body.Bytes(), &archived); err != nil {
		t.Fatal(err)
	}
	if archived.Status != string(taskdomain.StatusCancelled) {
		t.Fatalf("status=%q", archived.Status)
	}

	deleteRec := doJSON(t, env.router, http.MethodDelete, "/api/v1/tasks/"+created.ID, auth, nil)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleteRec.Code, deleteRec.Body.String())
	}
	missing := doJSON(t, env.router, http.MethodGet, "/api/v1/tasks/"+created.ID, auth, nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("get after delete status=%d want 404", missing.Code)
	}
}

func TestEditTaskNotFoundReturns404(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}
	id := ids.NewTaskID().String()
	rec := doJSON(t, env.router, http.MethodPatch, "/api/v1/tasks/"+id, auth, map[string]any{
		"title": "ghost",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404 body=%s", rec.Code, rec.Body.String())
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
		ID        string `json:"id"`
		Name      string `json:"name"`
		Frequency string `json:"frequency"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Name != "бег" || created.Frequency != "daily" || created.ID == "" {
		t.Fatalf("created=%+v", created)
	}

	todayBefore := doJSON(t, env.router, http.MethodGet, "/api/v1/habits/today", auth, nil)
	if todayBefore.Code != http.StatusOK {
		t.Fatalf("today status=%d body=%s", todayBefore.Code, todayBefore.Body.String())
	}
	var listed struct {
		Habits []struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			TodayCompleted bool   `json:"today_completed"`
			Streak         int    `json:"streak"`
		} `json:"habits"`
	}
	if err := json.Unmarshal(todayBefore.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Habits) != 1 || listed.Habits[0].ID != created.ID || listed.Habits[0].TodayCompleted {
		t.Fatalf("today before track=%+v", listed.Habits)
	}

	trackRec := doJSON(t, env.router, http.MethodPost, "/api/v1/habits/"+created.ID+"/track", auth, nil)
	if trackRec.Code != http.StatusOK {
		t.Fatalf("track status=%d body=%s", trackRec.Code, trackRec.Body.String())
	}
	var tracked struct {
		Name   string `json:"name"`
		Streak int    `json:"streak"`
	}
	if err := json.Unmarshal(trackRec.Body.Bytes(), &tracked); err != nil {
		t.Fatal(err)
	}
	if tracked.Name != "бег" || tracked.Streak < 1 {
		t.Fatalf("tracked=%+v", tracked)
	}

	todayAfter := doJSON(t, env.router, http.MethodGet, "/api/v1/habits/today", auth, nil)
	if err := json.Unmarshal(todayAfter.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Habits) != 1 || !listed.Habits[0].TodayCompleted || listed.Habits[0].Streak < 1 {
		t.Fatalf("today after track=%+v", listed.Habits)
	}

	missing := doJSON(t, env.router, http.MethodPost, "/api/v1/habits/"+ids.NewHabitID().String()+"/track", auth, nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing track status=%d body=%s", missing.Code, missing.Body.String())
	}
}

func TestCalendarHTTPContract(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	startsAt := time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/calendar/events", auth, map[string]any{
		"title":     "встреча",
		"starts_at": startsAt,
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		StartsAt string `json:"starts_at"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Title != "встреча" || created.StartsAt == "" {
		t.Fatalf("created=%+v", created)
	}

	listRec := doJSON(t, env.router, http.MethodGet, "/api/v1/calendar/today", auth, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var listed struct {
		Events []struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			StartsAt string `json:"starts_at"`
		} `json:"events"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Events) != 1 || listed.Events[0].ID != created.ID || listed.Events[0].Title != "встреча" {
		t.Fatalf("listed=%+v", listed.Events)
	}

	bad := doJSON(t, env.router, http.MethodPost, "/api/v1/calendar/events", auth, map[string]any{
		"title": "x", "starts_at": "not-a-date",
	})
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("bad starts_at status=%d", bad.Code)
	}
}

func TestSettingsAndSpheresHTTPContract(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	getRec := doJSON(t, env.router, http.MethodGet, "/api/v1/settings", auth, nil)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get settings status=%d body=%s", getRec.Code, getRec.Body.String())
	}
	var settings struct {
		MorningReviewAt struct {
			Hour   int `json:"hour"`
			Minute int `json:"minute"`
		} `json:"morning_review_at"`
		EveningReviewAt struct {
			Hour   int `json:"hour"`
			Minute int `json:"minute"`
		} `json:"evening_review_at"`
		WeeklyReviewAt struct {
			Hour   int `json:"hour"`
			Minute int `json:"minute"`
		} `json:"weekly_review_at"`
		MonthlyReviewAt struct {
			Hour   int `json:"hour"`
			Minute int `json:"minute"`
		} `json:"monthly_review_at"`
		QuietHoursStart *struct {
			Hour   int `json:"hour"`
			Minute int `json:"minute"`
		} `json:"quiet_hours_start"`
		QuietHoursEnd *struct {
			Hour   int `json:"hour"`
			Minute int `json:"minute"`
		} `json:"quiet_hours_end"`
		Language string `json:"language"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &settings); err != nil {
		t.Fatal(err)
	}
	if settings.Language != "ru" || settings.MorningReviewAt.Hour != 8 || settings.EveningReviewAt.Hour != 21 {
		t.Fatalf("settings=%+v", settings)
	}
	if settings.QuietHoursStart != nil || settings.QuietHoursEnd != nil {
		t.Fatalf("expected null quiet hours, got start=%v end=%v", settings.QuietHoursStart, settings.QuietHoursEnd)
	}

	morning := doJSON(t, env.router, http.MethodPut, "/api/v1/settings/morning-review", auth, map[string]any{
		"hour": 7, "minute": 30,
	})
	if morning.Code != http.StatusOK {
		t.Fatalf("morning status=%d body=%s", morning.Code, morning.Body.String())
	}
	evening := doJSON(t, env.router, http.MethodPut, "/api/v1/settings/evening-review", auth, map[string]any{
		"hour": 22, "minute": 0,
	})
	if evening.Code != http.StatusOK {
		t.Fatalf("evening status=%d body=%s", evening.Code, evening.Body.String())
	}
	quiet := doJSON(t, env.router, http.MethodPut, "/api/v1/settings/quiet-hours", auth, map[string]any{
		"start_hour": 23, "start_minute": 0, "end_hour": 7, "end_minute": 0,
	})
	if quiet.Code != http.StatusOK {
		t.Fatalf("quiet status=%d body=%s", quiet.Code, quiet.Body.String())
	}

	getRec = doJSON(t, env.router, http.MethodGet, "/api/v1/settings", auth, nil)
	if err := json.Unmarshal(getRec.Body.Bytes(), &settings); err != nil {
		t.Fatal(err)
	}
	if settings.MorningReviewAt.Hour != 7 || settings.MorningReviewAt.Minute != 30 {
		t.Fatalf("morning after update=%+v", settings.MorningReviewAt)
	}
	if settings.QuietHoursStart == nil || settings.QuietHoursStart.Hour != 23 {
		t.Fatalf("quiet start=%v", settings.QuietHoursStart)
	}

	createSphere := doJSON(t, env.router, http.MethodPost, "/api/v1/settings/spheres", auth, map[string]any{
		"name": "Здоровье",
	})
	if createSphere.Code != http.StatusCreated {
		t.Fatalf("create sphere status=%d body=%s", createSphere.Code, createSphere.Body.String())
	}
	var sphere struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		SortOrder int32  `json:"sort_order"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(createSphere.Body.Bytes(), &sphere); err != nil {
		t.Fatal(err)
	}
	if sphere.ID == "" || sphere.Name != "Здоровье" || sphere.CreatedAt == "" {
		t.Fatalf("sphere=%+v", sphere)
	}

	updateSphere := doJSON(t, env.router, http.MethodPut, "/api/v1/settings/spheres/"+sphere.ID, auth, map[string]any{
		"name": "Здоровье+", "sort_order": 2,
	})
	if updateSphere.Code != http.StatusOK {
		t.Fatalf("update sphere status=%d body=%s", updateSphere.Code, updateSphere.Body.String())
	}

	listSpheres := doJSON(t, env.router, http.MethodGet, "/api/v1/settings/spheres", auth, nil)
	if listSpheres.Code != http.StatusOK {
		t.Fatalf("list spheres status=%d", listSpheres.Code)
	}
	var listed struct {
		Spheres []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			SortOrder int32  `json:"sort_order"`
		} `json:"spheres"`
	}
	if err := json.Unmarshal(listSpheres.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Spheres) != 1 || listed.Spheres[0].Name != "Здоровье+" || listed.Spheres[0].SortOrder != 2 {
		t.Fatalf("listed spheres=%+v", listed.Spheres)
	}

	deleteSphere := doJSON(t, env.router, http.MethodDelete, "/api/v1/settings/spheres/"+sphere.ID, auth, nil)
	if deleteSphere.Code != http.StatusOK {
		t.Fatalf("delete sphere status=%d body=%s", deleteSphere.Code, deleteSphere.Body.String())
	}
	var deleted struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(deleteSphere.Body.Bytes(), &deleted); err != nil {
		t.Fatal(err)
	}
	if deleted.ID != sphere.ID {
		t.Fatalf("deleted=%+v", deleted)
	}

	missing := doJSON(t, env.router, http.MethodDelete, "/api/v1/settings/spheres/"+ids.NewSphereID().String(), auth, nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing delete status=%d", missing.Code)
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

type fakeOverviewStore struct {
	income     int64
	expense    int64
	categories []financeapp.CategoryExpenseTotal
}

func (s *fakeOverviewStore) SumIncomeBetween(context.Context, ids.UserID, time.Time, time.Time) (int64, error) {
	return s.income, nil
}

func (s *fakeOverviewStore) SumExpenseBetween(context.Context, ids.UserID, time.Time, time.Time) (int64, error) {
	return s.expense, nil
}

func (s *fakeOverviewStore) SumExpensesByCategoryBetween(context.Context, ids.UserID, time.Time, time.Time) ([]financeapp.CategoryExpenseTotal, error) {
	return s.categories, nil
}

func TestFinanceOverviewHTTP(t *testing.T) {
	t.Parallel()
	user, err := domain.NewUser(testTelegram, "API User", "UTC", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := auth.NewTokenService(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	users := &stubUserRepo{user: user}
	store := &fakeOverviewStore{
		income:  10_000,
		expense: 5_000,
		categories: []financeapp.CategoryExpenseTotal{
			{Name: "Еда", AmountCents: 3_000},
			{Name: "Транспорт", AmountCents: 2_000},
		},
	}
	rt := api.NewRouter(api.Deps{
		Log:             slog.Default(),
		APIKey:          testAPIKey,
		Tokens:          tokens,
		GetUser:         identityapp.NewGetUserByTelegram(users),
		FinanceOverview: financeapp.NewFinanceOverview(store, fakeTZ{}),
	})
	r := chi.NewRouter()
	rt.Mount(r)
	env := testEnv{user: user, router: r}
	token := issueToken(t, env)
	authHdr := map[string]string{"Authorization": "Bearer " + token}

	okRec := doJSON(t, env.router, http.MethodGet, "/api/v1/finance/overview?period=2026-07", authHdr, nil)
	if okRec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", okRec.Code, okRec.Body.String())
	}
	var body struct {
		PeriodLabel  string  `json:"period_label"`
		IncomeCents  int64   `json:"income_cents"`
		ExpenseCents int64   `json:"expense_cents"`
		NetCents     int64   `json:"net_cents"`
		Currency     string  `json:"currency"`
		Categories   []struct {
			Name        string  `json:"name"`
			AmountCents int64   `json:"amount_cents"`
			Percent     float64 `json:"percent"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(okRec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.PeriodLabel != "июль 2026" || body.Currency != "RUB" || body.NetCents != 5_000 {
		t.Fatalf("body=%+v", body)
	}
	if len(body.Categories) != 2 || body.Categories[0].Percent != 60 {
		t.Fatalf("categories=%+v", body.Categories)
	}

	badRec := doJSON(t, env.router, http.MethodGet, "/api/v1/finance/overview?period=2026-7", authHdr, nil)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid period status=%d", badRec.Code)
	}

	emptyRec := doJSON(t, env.router, http.MethodGet, "/api/v1/finance/overview", authHdr, nil)
	if emptyRec.Code != http.StatusBadRequest {
		t.Fatalf("missing period status=%d", emptyRec.Code)
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

	missing := doJSON(t, env.router, http.MethodDelete, "/api/v1/notes/"+created.ID, auth, nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("delete missing status=%d", missing.Code)
	}
	badID := doJSON(t, env.router, http.MethodDelete, "/api/v1/notes/not-a-uuid", auth, nil)
	if badID.Code != http.StatusBadRequest {
		t.Fatalf("delete bad id status=%d", badID.Code)
	}
}

func TestDebtPayHTTPContract(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/finance/debts", auth, map[string]any{
		"creditor":     "банку",
		"amount_cents": int64(1_000_000),
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID             string `json:"id"`
		AmountCents    int64  `json:"amount_cents"`
		PaidCents      int64  `json:"paid_cents"`
		RemainingCents int64  `json:"remaining_cents"`
		Currency       string `json:"currency"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Currency != "RUB" || created.RemainingCents != 1_000_000 {
		t.Fatalf("created=%+v", created)
	}

	payRec := doJSON(t, env.router, http.MethodPost, "/api/v1/finance/debts/"+created.ID+"/pay", auth, map[string]any{
		"amount_cents": int64(250_000),
	})
	if payRec.Code != http.StatusOK {
		t.Fatalf("pay status=%d body=%s", payRec.Code, payRec.Body.String())
	}
	var paid struct {
		PaidCents      int64 `json:"paid_cents"`
		RemainingCents int64 `json:"remaining_cents"`
	}
	if err := json.Unmarshal(payRec.Body.Bytes(), &paid); err != nil {
		t.Fatal(err)
	}
	if paid.PaidCents != 250_000 || paid.RemainingCents != 750_000 {
		t.Fatalf("paid=%+v", paid)
	}

	missing := doJSON(t, env.router, http.MethodPost, "/api/v1/finance/debts/"+uuid.Must(uuid.NewV7()).String()+"/pay", auth, map[string]any{
		"amount_cents": int64(1),
	})
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing debt status=%d", missing.Code)
	}
	badAmt := doJSON(t, env.router, http.MethodPost, "/api/v1/finance/debts/"+created.ID+"/pay", auth, map[string]any{
		"amount_cents": int64(0),
	})
	if badAmt.Code != http.StatusBadRequest {
		t.Fatalf("bad amount status=%d", badAmt.Code)
	}
}

func TestCareerHTTPContract(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	cRec := doJSON(t, env.router, http.MethodPost, "/api/v1/career/contacts", auth, map[string]any{
		"name": "Ada", "company": "Analytical", "role": "Engineer", "notes": "hi",
	})
	if cRec.Code != http.StatusCreated {
		t.Fatalf("create contact status=%d body=%s", cRec.Code, cRec.Body.String())
	}
	var contact struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Company   string `json:"company"`
		Role      string `json:"role"`
		Notes     string `json:"notes"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(cRec.Body.Bytes(), &contact); err != nil {
		t.Fatal(err)
	}
	if contact.Name != "Ada" || contact.Company != "Analytical" || contact.CreatedAt == "" {
		t.Fatalf("contact=%+v", contact)
	}

	listC := doJSON(t, env.router, http.MethodGet, "/api/v1/career/contacts", auth, nil)
	if listC.Code != http.StatusOK {
		t.Fatalf("list contacts status=%d", listC.Code)
	}
	var listedC struct {
		Contacts []struct {
			Name string `json:"name"`
		} `json:"contacts"`
	}
	if err := json.Unmarshal(listC.Body.Bytes(), &listedC); err != nil {
		t.Fatal(err)
	}
	if len(listedC.Contacts) != 1 {
		t.Fatalf("contacts=%+v", listedC.Contacts)
	}

	delC := doJSON(t, env.router, http.MethodDelete, "/api/v1/career/contacts/"+contact.ID, auth, nil)
	if delC.Code != http.StatusOK {
		t.Fatalf("delete contact status=%d", delC.Code)
	}
	missingC := doJSON(t, env.router, http.MethodDelete, "/api/v1/career/contacts/"+contact.ID, auth, nil)
	if missingC.Code != http.StatusNotFound {
		t.Fatalf("missing contact status=%d", missingC.Code)
	}

	sRec := doJSON(t, env.router, http.MethodPost, "/api/v1/career/skills", auth, map[string]any{
		"name": "Go", "level": "advanced",
	})
	if sRec.Code != http.StatusCreated {
		t.Fatalf("create skill status=%d body=%s", sRec.Code, sRec.Body.String())
	}
	var skill struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level string `json:"level"`
	}
	if err := json.Unmarshal(sRec.Body.Bytes(), &skill); err != nil {
		t.Fatal(err)
	}
	if skill.Name != "Go" || skill.Level != "advanced" {
		t.Fatalf("skill=%+v", skill)
	}
	listS := doJSON(t, env.router, http.MethodGet, "/api/v1/career/skills?q=go", auth, nil)
	if listS.Code != http.StatusOK {
		t.Fatalf("search skills status=%d", listS.Code)
	}
	delS := doJSON(t, env.router, http.MethodDelete, "/api/v1/career/skills/"+skill.ID, auth, nil)
	if delS.Code != http.StatusOK {
		t.Fatalf("delete skill status=%d", delS.Code)
	}
	missingS := doJSON(t, env.router, http.MethodDelete, "/api/v1/career/skills/"+skill.ID, auth, nil)
	if missingS.Code != http.StatusNotFound {
		t.Fatalf("missing skill status=%d", missingS.Code)
	}
}

func TestHealthHTTPContract(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	emptyW := doJSON(t, env.router, http.MethodGet, "/api/v1/health/weight/latest", auth, nil)
	if emptyW.Code != http.StatusNotFound {
		t.Fatalf("empty weight status=%d", emptyW.Code)
	}
	emptyS := doJSON(t, env.router, http.MethodGet, "/api/v1/health/steps/latest", auth, nil)
	if emptyS.Code != http.StatusNotFound {
		t.Fatalf("empty steps status=%d", emptyS.Code)
	}
	emptySleep := doJSON(t, env.router, http.MethodGet, "/api/v1/health/sleep/latest", auth, nil)
	if emptySleep.Code != http.StatusNotFound {
		t.Fatalf("empty sleep status=%d", emptySleep.Code)
	}

	wRec := doJSON(t, env.router, http.MethodPost, "/api/v1/health/weight", auth, map[string]any{
		"weight_kg": 72.5,
	})
	if wRec.Code != http.StatusCreated {
		t.Fatalf("record weight status=%d body=%s", wRec.Code, wRec.Body.String())
	}
	var weight struct {
		ID       string  `json:"id"`
		WeightKg float64 `json:"weight_kg"`
		LoggedAt string  `json:"logged_at"`
	}
	if err := json.Unmarshal(wRec.Body.Bytes(), &weight); err != nil {
		t.Fatal(err)
	}
	if weight.WeightKg != 72.5 || weight.LoggedAt == "" {
		t.Fatalf("weight=%+v", weight)
	}
	latestW := doJSON(t, env.router, http.MethodGet, "/api/v1/health/weight/latest", auth, nil)
	if latestW.Code != http.StatusOK {
		t.Fatalf("latest weight status=%d", latestW.Code)
	}

	stRec := doJSON(t, env.router, http.MethodPost, "/api/v1/health/steps", auth, map[string]any{
		"steps": 8000,
	})
	if stRec.Code != http.StatusCreated {
		t.Fatalf("record steps status=%d body=%s", stRec.Code, stRec.Body.String())
	}
	var steps struct {
		Steps int32 `json:"steps"`
	}
	if err := json.Unmarshal(stRec.Body.Bytes(), &steps); err != nil {
		t.Fatal(err)
	}
	if steps.Steps != 8000 {
		t.Fatalf("steps=%+v", steps)
	}

	slRec := doJSON(t, env.router, http.MethodPost, "/api/v1/health/sleep", auth, map[string]any{
		"duration_hours": 7.5,
	})
	if slRec.Code != http.StatusCreated {
		t.Fatalf("record sleep status=%d body=%s", slRec.Code, slRec.Body.String())
	}
	var sleep struct {
		DurationMinutes int32   `json:"duration_minutes"`
		DurationHours   float64 `json:"duration_hours"`
	}
	if err := json.Unmarshal(slRec.Body.Bytes(), &sleep); err != nil {
		t.Fatal(err)
	}
	if sleep.DurationMinutes != 450 || sleep.DurationHours != 7.5 {
		t.Fatalf("sleep=%+v", sleep)
	}

	badW := doJSON(t, env.router, http.MethodPost, "/api/v1/health/weight", auth, map[string]any{
		"weight_kg": 0,
	})
	if badW.Code != http.StatusBadRequest {
		t.Fatalf("bad weight status=%d", badW.Code)
	}
}

func TestRemindersHTTPContract(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	fireAt := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	createRec := doJSON(t, env.router, http.MethodPost, "/api/v1/reminders", auth, map[string]any{
		"message": "pill",
		"fire_at": fireAt,
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var created struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		FireAt  string `json:"fire_at"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Message != "pill" || created.Status != "pending" || created.FireAt == "" {
		t.Fatalf("create must return ReminderDTO with id, got %+v body=%s", created, createRec.Body.String())
	}

	listRec := doJSON(t, env.router, http.MethodGet, "/api/v1/reminders", auth, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d", listRec.Code)
	}
	var listed struct {
		Reminders []struct {
			ID      string `json:"id"`
			Message string `json:"message"`
			FireAt  string `json:"fire_at"`
			Status  string `json:"status"`
		} `json:"reminders"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Reminders) != 1 || listed.Reminders[0].Message != "pill" || listed.Reminders[0].Status != "pending" {
		t.Fatalf("listed=%+v", listed.Reminders)
	}
	if listed.Reminders[0].ID != created.ID {
		t.Fatalf("list id=%q != create id=%q", listed.Reminders[0].ID, created.ID)
	}

	delRec := doJSON(t, env.router, http.MethodDelete, "/api/v1/reminders/"+created.ID, auth, nil)
	if delRec.Code != http.StatusOK {
		t.Fatalf("cancel by create id status=%d body=%s", delRec.Code, delRec.Body.String())
	}
	var cancelled struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(delRec.Body.Bytes(), &cancelled); err != nil {
		t.Fatal(err)
	}
	if cancelled.ID != created.ID || cancelled.Status != "cancelled" {
		t.Fatalf("cancel response=%+v", cancelled)
	}
	missing := doJSON(t, env.router, http.MethodDelete, "/api/v1/reminders/"+created.ID, auth, nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing cancel status=%d", missing.Code)
	}

	past := doJSON(t, env.router, http.MethodPost, "/api/v1/reminders", auth, map[string]any{
		"message": "late",
		"fire_at": time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
	})
	if past.Code != http.StatusBadRequest {
		t.Fatalf("past fire_at status=%d", past.Code)
	}
}

func TestAnalyticsSummaryHTTPContract(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	auth := map[string]string{"Authorization": "Bearer " + token}

	rec := doJSON(t, env.router, http.MethodGet, "/api/v1/analytics/summary", auth, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		PeriodLabel      string `json:"period_label"`
		TasksCreated     int64  `json:"tasks_created"`
		TasksCompleted   int64  `json:"tasks_completed"`
		CompletionRate   int    `json:"completion_rate"`
		OpenTasks        int64  `json:"open_tasks"`
		HabitConsistency int    `json:"habit_consistency"`
		HabitCompletions int64  `json:"habit_completions"`
		HabitCount       int64  `json:"habit_count"`
		Projects         []struct {
			Title   string `json:"title"`
			Percent string `json:"percent"`
		} `json:"projects"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.CompletionRate != 70 || body.HabitConsistency != 80 {
		t.Fatalf("percents=%d/%d want ints 70/80", body.CompletionRate, body.HabitConsistency)
	}
	if len(body.Projects) != 1 || body.Projects[0].Title != "LifeOS" || body.Projects[0].Percent != "42%" {
		t.Fatalf("projects=%+v (expect lowercase title/percent)", body.Projects)
	}

	// Empty projects must serialize as [] not null.
	rt := api.NewRouter(api.Deps{
		Log:     slog.Default(),
		APIKey:  testAPIKey,
		Tokens:  mustTokens(t),
		GetUser: identityapp.NewGetUserByTelegram(&stubUserRepo{user: env.user}),
		Analytics: fakeAnalytics{summary: query.ProductivitySummary{
			PeriodLabel: "июль 2026",
			Projects:    nil,
		}},
	})
	r := chi.NewRouter()
	rt.Mount(r)
	token2 := issueToken(t, testEnv{user: env.user, router: r})
	emptyRec := doJSON(t, r, http.MethodGet, "/api/v1/analytics/summary", map[string]string{
		"Authorization": "Bearer " + token2,
	}, nil)
	if emptyRec.Code != http.StatusOK {
		t.Fatalf("empty projects status=%d", emptyRec.Code)
	}
	raw := emptyRec.Body.String()
	if !strings.Contains(raw, `"projects":[]`) {
		t.Fatalf("projects must be empty array, body=%s", raw)
	}
}

func mustTokens(t *testing.T) *auth.TokenService {
	t.Helper()
	tokens, err := auth.NewTokenService(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return tokens
}
