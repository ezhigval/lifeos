package cmd

import (
	"fmt"
	"log/slog"
	"time"

	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/platform/auth"
	"github.com/valentinezhov/lifeos/internal/platform/config"
	"github.com/valentinezhov/lifeos/internal/transport/http/api"
)

func (rt *runtime) apiRouter(cfg config.Config, log *slog.Logger) (*api.Router, error) {
	if cfg.JWTSecret == "" {
		return nil, nil
	}
	tokens, err := auth.NewTokenService(cfg.JWTSecret, time.Duration(cfg.JWTTTLHours)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("jwt: %w", err)
	}
	return api.NewRouter(api.Deps{
		Log:              log,
		APIKey:           cfg.APIKey,
		BotToken:         cfg.TelegramBotToken,
		WebAppAuthTTL:    time.Duration(cfg.WebAppAuthTTLHours) * time.Hour,
		Tokens:           tokens,
		GetUser:          identityapp.NewGetUserByTelegram(rt.users),
		EnsureUser:       rt.ensureUser,
		ListToday:        rt.listToday,
		CreateTask:       rt.createTask,
		Complete:         rt.completeTask,
		ReopenTask:       rt.reopenTask,
		CancelTask:       rt.cancelTask,
		EditTask:         rt.editTask,
		RescheduleTask:   rt.rescheduleTask,
		ListByTag:        rt.listTasksByTag,
		ListDueBetween:   rt.listDueBetween,
		GetTask:          rt.getTask,
		UpdateTask:       rt.updateTask,
		ArchiveTask:      rt.archiveTask,
		DeleteTask:       rt.deleteTask,
		ProjectProg:      rt.projectProg,
		Review:           rt.review,
		Priorities:       rt.priorities,
		Analytics:        rt.analytics,
		RecordIncome:     rt.recordIncome,
		RecordExpense:    rt.recordExpense,
		ListDebts:        rt.listDebts,
		CreateDebt:       rt.createDebt,
		PayDebt:          rt.payDebt,
		ListFinancePlan:  rt.listFinancePlan,
		CreatePlanned:    rt.createPlanned,
		DeletePlanned:    rt.deletePlanned,
		CashFlow:         rt.cashFlow,
		FinanceOverview:  rt.financeOverview,
		ListHabits:       rt.listHabits,
		CreateHabit:      rt.createHabit,
		TrackHabit:       rt.trackHabit,
		ScheduleReminder: rt.reminder,
		ListReminders:    rt.listReminders,
		CancelReminder:   rt.cancelReminder,
		CreateNote:       rt.createNote,
		ListNotes:        rt.listNotes,
		SearchNotes:      rt.searchNotes,
		GetNote:          rt.getNote,
		UpdateNote:       rt.updateNote,
		DeleteNote:       rt.deleteNote,
		CreateContact:    rt.createContact,
		ListContacts:     rt.listContacts,
		SearchContacts:   rt.searchContacts,
		DeleteContact:    rt.deleteContact,
		CreateSkill:      rt.createSkill,
		ListSkills:       rt.listSkills,
		SearchSkills:     rt.searchSkills,
		DeleteSkill:      rt.deleteSkill,
		CreateSphere:     rt.createSphere,
		ListSpheres:      rt.listSpheres,
		UpdateSphere:     rt.updateSphere,
		DeleteSphere:     rt.deleteSphere,
		RecordWeight:     rt.recordWeight,
		GetLatestWeight:  rt.latestWeight,
		ListWeights:      rt.listWeights,
		RecordSteps:      rt.recordSteps,
		GetLatestSteps:   rt.latestSteps,
		ListSteps:        rt.listSteps,
		RecordSleep:      rt.recordSleep,
		GetLatestSleep:   rt.latestSleep,
		ListSleep:        rt.listSleep,
		ListCalendar:     rt.listCalendar,
		CreateEvent:      rt.createEvent,
		ListProjects:     rt.listProjects,
		CreateProject:    rt.createProject,
		ListProjectTasks: rt.listProjectTasks,
		ArchiveProject:   rt.archiveProject,
		GetSettings:      rt.getSettings,
		UpdateMorning:    rt.updateMorning,
		UpdateEvening:    rt.updateEvening,
		UpdateQuiet:      rt.updateQuiet,
	}), nil
}
