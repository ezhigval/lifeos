package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	identityinfra "github.com/valentinezhov/lifeos/internal/identity/infra"
	"github.com/valentinezhov/lifeos/internal/platform/config"
	"github.com/valentinezhov/lifeos/internal/platform/db"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/platform/logging"
	"github.com/valentinezhov/lifeos/internal/platform/pgconv"
	"github.com/valentinezhov/lifeos/internal/platform/postgres"
	"github.com/valentinezhov/lifeos/internal/platform/timeutil"
	projectsinfra "github.com/valentinezhov/lifeos/internal/projects/infra"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	"github.com/valentinezhov/lifeos/internal/tasks/domain"
	tasksinfra "github.com/valentinezhov/lifeos/internal/tasks/infra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task debugging commands",
}

var taskCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a task for the seed user",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		exitOnError(runTaskCreate(strings.Join(args, " ")))
	},
}

var taskListTodayCmd = &cobra.Command{
	Use:   "list-today",
	Short: "List tasks due today for the seed user",
	Run: func(_ *cobra.Command, _ []string) {
		exitOnError(runTaskListToday())
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete [task-id]",
	Short: "Complete a task by id",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		exitOnError(runTaskComplete(args[0]))
	},
}

func init() {
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListTodayCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	rootCmd.AddCommand(taskCmd)
}

func runTaskCreate(title string) error {
	deps, err := newTaskDeps()
	if err != nil {
		return err
	}
	defer deps.pool.Close()

	uc := tasksapp.NewCreateTask(deps.store, deps.publisher, deps.transactor, projectsinfra.NewProjectReader(deps.pool.Pool))

	tz, err := deps.tzReader.Timezone(deps.ctx, deps.userID)
	if err != nil {
		return err
	}
	today, err := timeutil.DateInTimezone(time.Now().UTC(), tz)
	if err != nil {
		return err
	}

	dto, err := uc.Execute(deps.ctx, tasksapp.CreateTaskInput{
		UserID:   deps.userID,
		Title:    title,
		Priority: domain.PriorityMedium,
		DueDate:  &today,
		Source:   events.SourceCLI,
	})
	if err != nil {
		return err
	}

	fmt.Printf("created task %s: %s\n", dto.ID, dto.Title)
	return nil
}

func runTaskListToday() error {
	deps, err := newTaskDeps()
	if err != nil {
		return err
	}
	defer deps.pool.Close()

	uc := tasksapp.NewListTasksToday(deps.store, deps.tzReader)
	items, err := uc.Execute(deps.ctx, deps.userID)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("no tasks for today")
		return nil
	}

	for _, item := range items {
		fmt.Printf("- [%s] %s (%s)\n", item.Status, item.Title, item.Priority)
	}
	return nil
}

func runTaskComplete(taskID string) error {
	deps, err := newTaskDeps()
	if err != nil {
		return err
	}
	defer deps.pool.Close()

	id, err := ids.ParseTaskID(taskID)
	if err != nil {
		return err
	}

	uc := tasksapp.NewCompleteTask(deps.store, deps.publisher, deps.transactor)
	dto, err := uc.Execute(deps.ctx, tasksapp.CompleteTaskInput{
		UserID: deps.userID,
		TaskID: id,
		Source: events.SourceCLI,
	})
	if err != nil {
		return err
	}

	fmt.Printf("completed task %s: %s\n", dto.ID, dto.Title)
	return nil
}

type taskDeps struct {
	ctx        context.Context
	pool       *postgres.Pool
	userID     ids.UserID
	store      *tasksinfra.Repository
	publisher  *events.Publisher
	transactor *postgres.Transactor
	tzReader   *identityinfra.TimezoneReader
}

func newTaskDeps() (*taskDeps, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if cfg.SeedTelegramID <= 0 {
		return nil, fmt.Errorf("set LIFEOS_SEED_TELEGRAM_ID in .env")
	}

	_ = logging.New(cfg.LogLevel, cfg.LogFormat)
	ctx := context.Background()

	pool, err := postgres.New(ctx, cfg.DatabaseURL, false)
	if err != nil {
		return nil, err
	}

	queries := db.New(pool.Pool)
	user, err := queries.GetUserByTelegramID(ctx, cfg.SeedTelegramID)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("load seed user: %w", err)
	}

	return &taskDeps{
		ctx:        ctx,
		pool:       pool,
		userID:     pgconv.FromUserID(user.ID),
		store:      tasksinfra.NewRepository(pool.Pool),
		publisher:  events.NewPublisher(pool.Pool),
		transactor: postgres.NewTransactor(pool.Pool),
		tzReader:   identityinfra.NewTimezoneReader(pool.Pool),
	}, nil
}
