//go:build integration

package infra_test

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

	"github.com/valentinezhov/lifeos/internal/identity/domain"
	identityinfra "github.com/valentinezhov/lifeos/internal/identity/infra"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	platformmigrate "github.com/valentinezhov/lifeos/internal/platform/migrate"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
	tasksinfra "github.com/valentinezhov/lifeos/internal/tasks/infra"
)

func TestTaskRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := startPostgres(t, ctx)
	defer cleanup()

	userRepo := identityinfra.NewRepository(pool)
	now := time.Now().UTC()
	user, err := domain.NewUser(999001, "Test", "Europe/Moscow", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := userRepo.Upsert(ctx, user); err != nil {
		t.Fatal(err)
	}

	repo := tasksinfra.NewRepository(pool)
	publisher := events.NewPublisher(pool)
	transactor := platformpostgres.NewTransactor(pool)

	create := tasksapp.NewCreateTask(repo, publisher, transactor, nil)
	dto, err := create.Execute(ctx, tasksapp.CreateTaskInput{
		UserID:   user.ID,
		Title:    "integration task",
		Priority: taskdomain.PriorityHigh,
		Source:   events.SourceCLI,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	got, err := repo.GetByID(ctx, user.ID, dto.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if got.Title != "integration task" {
		t.Fatalf("title = %q", got.Title)
	}

	complete := tasksapp.NewCompleteTask(repo, publisher, transactor)
	done, err := complete.Execute(ctx, tasksapp.CompleteTaskInput{
		UserID: user.ID,
		TaskID: dto.ID,
		Source: events.SourceCLI,
	})
	if err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if done.Status != taskdomain.StatusDone {
		t.Fatalf("status = %s", done.Status)
	}
}

func startPostgres(t *testing.T, ctx context.Context) (*pgxpool.Pool, func()) {
	t.Helper()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("lifeos"),
		postgres.WithUsername("lifeos"),
		postgres.WithPassword("lifeos"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := platformmigrate.Up(ctx, connStr, migrationsDir); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	cleanup := func() {
		pool.Close()
		_ = container.Terminate(ctx)
	}
	return pool, cleanup
}
