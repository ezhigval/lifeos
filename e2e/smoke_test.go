//go:build integration

package e2e_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
	projectsinfra "github.com/valentinezhov/lifeos/internal/projects/infra"
	"github.com/valentinezhov/lifeos/internal/identity/domain"
	identityinfra "github.com/valentinezhov/lifeos/internal/identity/infra"
	notifapp "github.com/valentinezhov/lifeos/internal/notifications/app"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	platformmigrate "github.com/valentinezhov/lifeos/internal/platform/migrate"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
	tasksinfra "github.com/valentinezhov/lifeos/internal/tasks/infra"
)

func TestSmokeCreateTaskAndReminder(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := startPG(t, ctx)
	defer cleanup()

	userRepo := identityinfra.NewRepository(pool)
	user, _ := domain.NewUser(4242, "Smoke", "Europe/Moscow", time.Now().UTC())
	if err := userRepo.Upsert(ctx, user); err != nil {
		t.Fatal(err)
	}

	taskRepo := tasksinfra.NewRepository(pool)
	pub := events.NewPublisher(pool)
	tx := platformpostgres.NewTransactor(pool)
	create := tasksapp.NewCreateTask(taskRepo, pub, tx, projectsinfra.NewProjectReader(pool))

	dto, err := create.Execute(ctx, tasksapp.CreateTaskInput{
		UserID: user.ID, Title: "smoke task", Priority: taskdomain.PriorityHigh, Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dto.Title != "smoke task" {
		t.Fatalf("got %q", dto.Title)
	}

	reminder := notifapp.NewScheduleReminder(notifapp.NewJobStore(pool))
	if err := reminder.Execute(ctx, notifapp.ScheduleReminderInput{
		UserID: user.ID, Message: "ping", FireAt: time.Now().Add(time.Minute),
	}); err != nil {
		t.Fatal(err)
	}

	r := rulebased.NewResolver()
	intent, err := r.Resolve(ctx, ai.ResolveInput{Text: "добавь задачу e2e"})
	if err != nil || intent.Type != ai.IntentTaskCreate {
		t.Fatalf("intent: %+v err=%v", intent, err)
	}
}

func startPG(t *testing.T, ctx context.Context) (*pgxpool.Pool, func()) {
	t.Helper()
	c, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("lifeos"),
		postgres.WithUsername("lifeos"),
		postgres.WithPassword("lifeos"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := c.ConnectionString(ctx, "sslmode=disable")
	migrationsDir := filepath.Join("..", "migrations")
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := platformmigrate.Up(ctx, conn, migrationsDir); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	return pool, func() {
		pool.Close()
		_ = c.Terminate(ctx)
	}
}
