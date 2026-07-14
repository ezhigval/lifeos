package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	financeapp "github.com/valentinezhov/lifeos/internal/finance/app"
	financeinfra "github.com/valentinezhov/lifeos/internal/finance/infra"
	habitsapp "github.com/valentinezhov/lifeos/internal/habits/app"
	habitsinfra "github.com/valentinezhov/lifeos/internal/habits/infra"
	healthinfra "github.com/valentinezhov/lifeos/internal/health/infra"
	healthapp "github.com/valentinezhov/lifeos/internal/health/app"
	knowledgeapp "github.com/valentinezhov/lifeos/internal/knowledge/app"
	knowledgeinfra "github.com/valentinezhov/lifeos/internal/knowledge/infra"
	careerapp "github.com/valentinezhov/lifeos/internal/career/app"
	careerinfra "github.com/valentinezhov/lifeos/internal/career/infra"
	calendarinfra "github.com/valentinezhov/lifeos/internal/calendar/infra"
	calendarapp "github.com/valentinezhov/lifeos/internal/calendar/app"
	projectsinfra "github.com/valentinezhov/lifeos/internal/projects/infra"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	identitydomain "github.com/valentinezhov/lifeos/internal/identity/domain"
	identityinfra "github.com/valentinezhov/lifeos/internal/identity/infra"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	notifinfra "github.com/valentinezhov/lifeos/internal/notifications/infra"
	planapp "github.com/valentinezhov/lifeos/internal/planning/app"
	planinfra "github.com/valentinezhov/lifeos/internal/planning/infra"
	"github.com/valentinezhov/lifeos/internal/platform/config"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/postgres"
	"github.com/valentinezhov/lifeos/internal/platform/scheduler"
	"github.com/valentinezhov/lifeos/internal/query"
	settingsapp "github.com/valentinezhov/lifeos/internal/settings/app"
	settingsinfra "github.com/valentinezhov/lifeos/internal/settings/infra"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	spheresinfra "github.com/valentinezhov/lifeos/internal/spheres/infra"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	tasksinfra "github.com/valentinezhov/lifeos/internal/tasks/infra"
	tg "github.com/valentinezhov/lifeos/internal/transport/telegram"
	tginfra "github.com/valentinezhov/lifeos/internal/transport/telegram/infra"
)

type runtime struct {
	cfg      config.Config
	log      *slog.Logger
	pool     *pgxpool.Pool
	handler  *tg.MessageHandler
	poller   *tg.Poller
	sched    *scheduler.Scheduler
	tgClient *tg.Client
	notifier *notifinfra.TelegramNotifier
	review   *query.Review
	quiet    *settingsinfra.QuietHours
	reminder *notifapp.ScheduleReminder
	listReminders *notifapp.ListReminders
	cancelReminder *notifapp.CancelReminder
	settings *settingsinfra.Repository
	users    *identityinfra.Repository

	listToday  *tasksapp.ListTasksToday
	createTask *tasksapp.CreateTask
	completeTask *tasksapp.CompleteTask
	projectProg  *projectsapp.GetProjectProgress
	priorities *query.GetTopPriorities
	analytics  *query.GetProductivitySummary
	recordIncome *financeapp.RecordIncome
	recordExpense *financeapp.RecordExpense
	listDebts  *financeapp.ListDebts
	createDebt *financeapp.CreateDebt
	payDebt    *financeapp.PayDebt
	cashFlow   *financeapp.CashFlowSummary
	listHabits *habitsapp.ListHabitsToday
	createHabit *habitsapp.CreateHabit
	trackHabit *habitsapp.TrackHabit
	listCalendar *calendarapp.ListEventsToday
	createEvent  *calendarapp.CreateEvent
	listProjects *projectsapp.ListProjects
	createProject *projectsapp.CreateProject
	archiveProject *projectsapp.ArchiveProject
	listProjectTasks *tasksapp.ListTasksByProject
	getSettings  *settingsapp.GetSettings
	updateMorning *settingsapp.UpdateMorningReview
	updateEvening *settingsapp.UpdateEveningReview
	updateQuiet  *settingsapp.UpdateQuietHours
	createNote   *knowledgeapp.CreateNote
	listNotes    *knowledgeapp.ListNotes
	searchNotes  *knowledgeapp.SearchNotes
	deleteNote   *knowledgeapp.DeleteNote
	recordWeight *healthapp.RecordWeight
	latestWeight *healthapp.GetLatestWeight
	listWeights  *healthapp.ListWeights
	recordSteps  *healthapp.RecordSteps
	latestSteps  *healthapp.GetLatestSteps
	listSteps    *healthapp.ListSteps
	recordSleep  *healthapp.RecordSleep
	latestSleep  *healthapp.GetLatestSleep
	listSleep    *healthapp.ListSleep
	createContact *careerapp.CreateContact
	listContacts  *careerapp.ListContacts
	searchContacts *careerapp.SearchContacts
	deleteContact *careerapp.DeleteContact
	createSkill   *careerapp.CreateSkill
	listSkills    *careerapp.ListSkills
	searchSkills  *careerapp.SearchSkills
	deleteSkill   *careerapp.DeleteSkill
	createSphere  *spheresapp.CreateSphere
	listSpheres   *spheresapp.ListSpheres
	updateSphere  *spheresapp.UpdateSphere
	deleteSphere  *spheresapp.DeleteSphere
	findSphere    *spheresapp.FindSphereByName
}

func newRuntime(_ context.Context, cfg config.Config, log *slog.Logger, pool *postgres.Pool) (*runtime, error) {
	p := pool.Pool
	transactor := postgres.NewTransactor(p)
	eventPub := events.NewPublisher(p)
	taskRepo := tasksinfra.NewRepository(p)
	userRepo := identityinfra.NewRepository(p)
	tzReader := identityinfra.NewTimezoneReader(p)
	settingsRepo := settingsinfra.NewRepository(p)
	reminder := notifapp.NewScheduleReminder(notifapp.NewJobStore(p))
	listReminders := notifapp.NewListReminders(p)
	cancelReminder := notifapp.NewCancelReminder(p)
	ensureSettings := settingsapp.NewEnsureDefaults(settingsRepo)

	createTask := tasksapp.NewCreateTask(taskRepo, eventPub, transactor, projectsinfra.NewProjectReader(p))
	completeTask := tasksapp.NewCompleteTask(taskRepo, eventPub, transactor)
	completeByTitle := tasksapp.NewCompleteTaskByTitle(taskRepo, completeTask)
	listToday := tasksapp.NewListTasksToday(taskRepo, tzReader)
	financeRepo := financeinfra.NewRepository(p)
	recordIncome := financeapp.NewRecordIncome(financeRepo, financeRepo, eventPub, transactor, tzReader)
	recordExpense := financeapp.NewRecordExpense(financeRepo, financeRepo, eventPub, transactor, tzReader)
	createDebt := financeapp.NewCreateDebt(financeRepo, eventPub, transactor)
	listDebts := financeapp.NewListDebts(financeRepo)
	payDebt := financeapp.NewPayDebt(financeRepo, eventPub, transactor)
	cashFlow := financeapp.NewCashFlowSummary(financeRepo, tzReader)
	habitRepo := habitsinfra.NewRepository(p)
	createHabit := habitsapp.NewCreateHabit(habitRepo, eventPub, transactor)
	trackHabit := habitsapp.NewTrackHabit(habitRepo, habitRepo, eventPub, transactor, tzReader)
	listHabits := habitsapp.NewListHabitsToday(habitRepo, habitRepo, tzReader)
	noteRepo := knowledgeinfra.NewRepository(p)
	createNote := knowledgeapp.NewCreateNote(noteRepo, eventPub, transactor)
	listNotes := knowledgeapp.NewListNotes(noteRepo)
	searchNotes := knowledgeapp.NewSearchNotes(noteRepo)
	deleteNote := knowledgeapp.NewDeleteNote(noteRepo, eventPub, transactor)
	careerRepo := careerinfra.NewRepository(p)
	createContact := careerapp.NewCreateContact(careerRepo, eventPub, transactor)
	listContacts := careerapp.NewListContacts(careerRepo)
	searchContacts := careerapp.NewSearchContacts(careerRepo)
	deleteContact := careerapp.NewDeleteContact(careerRepo, eventPub, transactor)
	createSkill := careerapp.NewCreateSkill(careerRepo, eventPub, transactor)
	listSkills := careerapp.NewListSkills(careerRepo)
	searchSkills := careerapp.NewSearchSkills(careerRepo)
	deleteSkill := careerapp.NewDeleteSkill(careerRepo, eventPub, transactor)
	sphereRepo := spheresinfra.NewRepository(p)
	createSphere := spheresapp.NewCreateSphere(sphereRepo, eventPub, transactor)
	listSpheres := spheresapp.NewListSpheres(sphereRepo)
	updateSphere := spheresapp.NewUpdateSphere(sphereRepo, eventPub, transactor)
	deleteSphere := spheresapp.NewDeleteSphere(sphereRepo, eventPub, transactor)
	findSphere := spheresapp.NewFindSphereByName(sphereRepo)
	ensureDefaultSpheres := spheresapp.NewEnsureDefaultSpheres(sphereRepo)
	userBootstrap := &userBootstrap{settings: ensureSettings, spheres: ensureDefaultSpheres}
	healthRepo := healthinfra.NewRepository(p)
	recordWeight := healthapp.NewRecordWeight(healthRepo, eventPub, transactor)
	latestWeight := healthapp.NewGetLatestWeight(healthRepo)
	listWeights := healthapp.NewListWeights(healthRepo)
	recordSteps := healthapp.NewRecordSteps(healthRepo, eventPub, transactor)
	latestSteps := healthapp.NewGetLatestSteps(healthRepo)
	listSteps := healthapp.NewListSteps(healthRepo)
	recordSleep := healthapp.NewRecordSleep(healthRepo, eventPub, transactor)
	latestSleep := healthapp.NewGetLatestSleep(healthRepo)
	listSleep := healthapp.NewListSleep(healthRepo)
	projectRepo := projectsinfra.NewRepository(p)
	createProject := projectsapp.NewCreateProject(projectRepo, eventPub, transactor)
	findProject := projectsapp.NewFindProjectByName(projectRepo)
	listProjects := projectsapp.NewListProjects(projectRepo)
	projectProgress := projectsapp.NewGetProjectProgress(projectRepo)
	archiveProject := projectsapp.NewArchiveProject(projectRepo, eventPub, transactor)
	listProjectTasks := tasksapp.NewListTasksByProject(taskRepo)
	calendarRepo := calendarinfra.NewRepository(p)
	createEvent := calendarapp.NewCreateEvent(calendarRepo, eventPub, transactor)
	listCalendar := calendarapp.NewListEventsToday(calendarRepo, tzReader)
	sessions := tginfra.NewSessions(p)
	reviewScheduler := notifapp.NewReviewScheduler(p)
	updateMorning := settingsapp.NewUpdateMorningReview(settingsRepo, reviewScheduler, tzReader.Timezone, settingsinfra.ReviewAt)
	updateEvening := settingsapp.NewUpdateEveningReview(settingsRepo, reviewScheduler, tzReader.Timezone, settingsinfra.ReviewAt)
	updateQuiet := settingsapp.NewUpdateQuietHours(settingsRepo)
	getSettings := settingsapp.NewGetSettings(settingsRepo)
	priorities := query.NewGetTopPriorities(p, tzReader)
	analytics := query.NewGetProductivitySummary(p, tzReader)
	planRepo := planinfra.NewRepository(p)
	setAvail := planapp.NewSetDayAvailability(planRepo, tzReader)
	triage := planapp.NewTriageOverloadedDay(taskRepo, tzReader)
	reschedule := planapp.NewRescheduleTasks(taskRepo, tzReader)

	rt := &runtime{
		cfg:      cfg,
		log:      log,
		pool:     p,
		review:   query.NewReview(p, tzReader, newAssistant(cfg, log)),
		quiet:    settingsinfra.NewQuietHours(p),
		reminder: reminder,
		listReminders: listReminders,
		cancelReminder: cancelReminder,
		settings: settingsRepo,
		users:    userRepo,
		listToday:  listToday,
		createTask: createTask,
		completeTask: completeTask,
		projectProg:  projectProgress,
		priorities: priorities,
		analytics:  analytics,
		recordIncome: recordIncome,
		recordExpense: recordExpense,
		listDebts:  listDebts,
		createDebt: createDebt,
		payDebt:    payDebt,
		cashFlow:   cashFlow,
		listHabits: listHabits,
		createHabit: createHabit,
		trackHabit: trackHabit,
		listCalendar: listCalendar,
		createEvent:  createEvent,
		listProjects: listProjects,
		createProject: createProject,
		archiveProject: archiveProject,
		listProjectTasks: listProjectTasks,
		getSettings:  getSettings,
		updateMorning: updateMorning,
		updateEvening: updateEvening,
		updateQuiet:  updateQuiet,
		createNote:   createNote,
		listNotes:    listNotes,
		searchNotes:  searchNotes,
		deleteNote:   deleteNote,
		recordWeight: recordWeight,
		latestWeight: latestWeight,
		listWeights:  listWeights,
		recordSteps:  recordSteps,
		latestSteps:  latestSteps,
		listSteps:    listSteps,
		recordSleep:  recordSleep,
		latestSleep:  latestSleep,
		listSleep:    listSleep,
		createContact: createContact,
		listContacts:  listContacts,
		searchContacts: searchContacts,
		deleteContact: deleteContact,
		createSkill:   createSkill,
		listSkills:    listSkills,
		searchSkills:  searchSkills,
		deleteSkill:   deleteSkill,
		createSphere:  createSphere,
		listSpheres:   listSpheres,
		updateSphere:  updateSphere,
		deleteSphere:  deleteSphere,
		findSphere:    findSphere,
	}

	bootstrapUserReviews := func(ctx context.Context, user identitydomain.User) error {
		return rt.bootstrapReviewsForUser(ctx, user)
	}
	ensureUser := identityapp.NewEnsureUserByTelegram(
		userRepo,
		userBootstrap,
		cfg.SeedTimezone,
		bootstrapUserReviews,
	)

	if cfg.TelegramBotToken != "" {
		client := tg.NewClient(cfg.TelegramBotToken)
		rt.tgClient = client
		rt.notifier = notifinfra.NewTelegramNotifier(client, log)
		rt.handler = tg.NewHandler(tg.Deps{
			Log:             log,
			Client:          client,
			Resolver:        newIntentResolver(cfg, log),
			Sessions:        sessions,
			EnsureUser:      ensureUser,
			Processed:       tginfra.NewProcessedUpdates(p),
			CreateTask:      createTask,
			CompleteTask:    completeTask,
			CompleteByTitle: completeByTitle,
			ListToday:       listToday,
			ProjectProg:     projectProgress,
			UpdateMorning:   updateMorning,
			UpdateEvening:   updateEvening,
			UpdateQuiet:     updateQuiet,
			Priorities:      priorities,
			Analytics:       analytics,
			Reminder:        reminder,
			ListReminders:   listReminders,
			CancelReminder:  cancelReminder,
			SetAvail:        setAvail,
			Triage:          triage,
			Reschedule:      reschedule,
			RecordIncome:    recordIncome,
			RecordExpense:   recordExpense,
			CreateDebt:      createDebt,
			PayDebt:         payDebt,
			ListDebts:       listDebts,
			CashFlow:        cashFlow,
			CreateHabit:     createHabit,
			TrackHabit:      trackHabit,
			ListHabits:      listHabits,
			CreateNote:      createNote,
			ListNotes:       listNotes,
			SearchNotes:     searchNotes,
			DeleteNote:      deleteNote,
			CreateContact:   createContact,
			ListContacts:    listContacts,
			SearchContacts:  searchContacts,
			DeleteContact:   deleteContact,
			CreateSkill:     createSkill,
			ListSkills:      listSkills,
			SearchSkills:    searchSkills,
			DeleteSkill:     deleteSkill,
			CreateSphere:    createSphere,
			ListSpheres:     listSpheres,
			UpdateSphere:    updateSphere,
			DeleteSphere:    deleteSphere,
			FindSphere:      findSphere,
			RecordWeight:    recordWeight,
			LatestWeight:    latestWeight,
			RecordSteps:     recordSteps,
			LatestSteps:     latestSteps,
			RecordSleep:     recordSleep,
			LatestSleep:     latestSleep,
			CreateProject:   createProject,
			FindProject:     findProject,
			ListProjects:    listProjects,
			ListProjectTasks: listProjectTasks,
			ArchiveProject:  archiveProject,
			CreateEvent:     createEvent,
			ListCalendar:    listCalendar,
			Review:          rt.review,
			TZReader:        tzReader,
		})
		rt.poller = tg.NewPoller(client, rt.handler, log)
	}

	rt.sched = scheduler.New(p, log)
	if rt.notifier != nil {
		rt.wireScheduler()
		if err := rt.bootstrapReviews(context.Background()); err != nil {
			log.Warn("bootstrap reviews failed", "error", err)
		}
	}
	return rt, nil
}

func (rt *runtime) wireScheduler() {
	rt.sched.Register("reminder", func(ctx context.Context, userID ids.UserID, payload json.RawMessage) error {
		if skip, _ := rt.quiet.InQuietPeriod(ctx, userID, time.Now().UTC()); skip {
			rt.log.Info("defer reminder due to quiet hours", "user_id", userID.String())
			return scheduler.Defer(30 * time.Minute)
		}
		var p struct{ Message string }
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		return rt.notifier.Send(ctx, rt.telegramChatID(ctx, userID), p.Message)
	})
	rt.sched.Register("morning_review", rt.periodicReviewHandler("morning_review"))
	rt.sched.Register("evening_review", rt.periodicReviewHandler("evening_review"))
	rt.sched.Register("weekly_review", rt.periodicReviewHandler("weekly_review"))
	rt.sched.Register("monthly_review", rt.periodicReviewHandler("monthly_review"))
}

func (rt *runtime) telegramChatID(ctx context.Context, userID ids.UserID) int64 {
	user, err := rt.users.GetByID(ctx, userID)
	if err != nil {
		rt.log.Warn("telegram chat id lookup failed", "user_id", userID.String(), "error", err)
		return 0
	}
	return user.TelegramID
}

func (rt *runtime) periodicReviewHandler(jobType string) scheduler.JobHandler {
	return func(ctx context.Context, userID ids.UserID, _ json.RawMessage) error {
		if skip, _ := rt.quiet.InQuietPeriod(ctx, userID, time.Now().UTC()); skip {
			rt.log.Info("defer review due to quiet hours", "user_id", userID.String(), "job_type", jobType)
			return scheduler.Defer(30 * time.Minute)
		}
		var text string
		var err error
		switch jobType {
		case "morning_review":
			text, err = rt.review.Morning(ctx, userID)
		case "evening_review":
			text, err = rt.review.Evening(ctx, userID)
		case "weekly_review":
			text, err = rt.review.Weekly(ctx, userID)
		case "monthly_review":
			text, err = rt.review.Monthly(ctx, userID, true)
		default:
			return fmt.Errorf("unknown review job: %s", jobType)
		}
		if err != nil {
			return err
		}
		if err := rt.notifier.Send(ctx, rt.telegramChatID(ctx, userID), text); err != nil {
			return err
		}
		return rt.scheduleNextPeriodic(ctx, userID, jobType)
	}
}

func (rt *runtime) scheduleNextPeriodic(ctx context.Context, userID ids.UserID, jobType string) error {
	tz, err := identityinfra.NewTimezoneReader(rt.pool).Timezone(ctx, userID)
	if err != nil {
		return err
	}
	settings, err := rt.settings.Get(ctx, userID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	var next time.Time
	switch jobType {
	case "morning_review":
		next, err = settingsinfra.ReviewAt(now, tz, settings.MorningReviewAt)
		if err == nil {
			next = next.Add(24 * time.Hour)
		}
	case "evening_review":
		next, err = settingsinfra.ReviewAt(now, tz, settings.EveningReviewAt)
		if err == nil {
			next = next.Add(24 * time.Hour)
		}
	case "weekly_review":
		next, err = settingsinfra.NextWeeklyReviewAt(now.Add(time.Minute), tz, settings.WeeklyReviewAt)
	case "monthly_review":
		next, err = settingsinfra.NextMonthlyReviewAt(now.Add(time.Minute), tz, settings.MonthlyReviewAt)
	default:
		return fmt.Errorf("unknown review job: %s", jobType)
	}
	if err != nil {
		return err
	}
	return rt.reminder.EnsureReview(ctx, userID, jobType, next)
}

func (rt *runtime) bootstrapReviews(ctx context.Context) error {
	users, err := rt.users.ListAll(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		if err := rt.bootstrapReviewsForUser(ctx, user); err != nil {
			rt.log.Warn("bootstrap reviews failed", "user_id", user.ID.String(), "error", err)
		}
	}
	return nil
}

func (rt *runtime) bootstrapReviewsForUser(ctx context.Context, user identitydomain.User) error {
	settings, err := rt.settings.Get(ctx, user.ID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	morning, err := settingsinfra.ReviewAt(now, user.Timezone, settings.MorningReviewAt)
	if err != nil {
		return err
	}
	evening, err := settingsinfra.ReviewAt(now, user.Timezone, settings.EveningReviewAt)
	if err != nil {
		return err
	}
	if err := rt.reminder.EnsureReview(ctx, user.ID, "morning_review", morning); err != nil {
		return err
	}
	if err := rt.reminder.EnsureReview(ctx, user.ID, "evening_review", evening); err != nil {
		return err
	}
	weekly, err := settingsinfra.NextWeeklyReviewAt(now, user.Timezone, settings.WeeklyReviewAt)
	if err != nil {
		return err
	}
	monthly, err := settingsinfra.NextMonthlyReviewAt(now, user.Timezone, settings.MonthlyReviewAt)
	if err != nil {
		return err
	}
	if err := rt.reminder.EnsureReview(ctx, user.ID, "weekly_review", weekly); err != nil {
		return err
	}
	return rt.reminder.EnsureReview(ctx, user.ID, "monthly_review", monthly)
}
