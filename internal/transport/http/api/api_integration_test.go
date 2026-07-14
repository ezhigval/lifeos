//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/valentinezhov/lifeos/internal/query"
	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/identity/domain"
	identityinfra "github.com/valentinezhov/lifeos/internal/identity/infra"
	projectsapp "github.com/valentinezhov/lifeos/internal/projects/app"
	projectsinfra "github.com/valentinezhov/lifeos/internal/projects/infra"
	spheresapp "github.com/valentinezhov/lifeos/internal/spheres/app"
	spheresinfra "github.com/valentinezhov/lifeos/internal/spheres/infra"
	"github.com/valentinezhov/lifeos/internal/platform/auth"
	"github.com/valentinezhov/lifeos/internal/platform/events"
	platformmigrate "github.com/valentinezhov/lifeos/internal/platform/migrate"
	platformpostgres "github.com/valentinezhov/lifeos/internal/platform/postgres"
	settingsapp "github.com/valentinezhov/lifeos/internal/settings/app"
	settingsinfra "github.com/valentinezhov/lifeos/internal/settings/infra"
	"github.com/valentinezhov/lifeos/internal/transport/http/api"
	tasksapp "github.com/valentinezhov/lifeos/internal/tasks/app"
	taskdomain "github.com/valentinezhov/lifeos/internal/tasks/domain"
	tasksinfra "github.com/valentinezhov/lifeos/internal/tasks/infra"
)

const (
	intAPIKey    = "integration-api-key"
	intJWTSecret = "integration-secret-key-32bytes!!"
	intTelegram  = int64(880088)
)

func TestAPIIntegrationProjectArchive(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := startPG(t, ctx)
	defer cleanup()

	userRepo := identityinfra.NewRepository(pool)
	user, err := domain.NewUser(intTelegram, "Integration", "UTC", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := userRepo.Upsert(ctx, user); err != nil {
		t.Fatal(err)
	}
	settingsRepo := settingsinfra.NewRepository(pool)
	if err := settingsapp.NewEnsureDefaults(settingsRepo).Execute(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	transactor := platformpostgres.NewTransactor(pool)
	eventPub := events.NewPublisher(pool)
	projectRepo := projectsinfra.NewRepository(pool)
	createProject := projectsapp.NewCreateProject(projectRepo, eventPub, transactor)
	archiveProject := projectsapp.NewArchiveProject(projectRepo, eventPub, transactor)

	tokens, err := auth.NewTokenService(intJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	rt := api.NewRouter(api.Deps{
		Log:            slog.Default(),
		APIKey:         intAPIKey,
		Tokens:         tokens,
		GetUser:        identityapp.NewGetUserByTelegram(userRepo),
		CreateProject:  createProject,
		ArchiveProject: archiveProject,
	})
	r := chi.NewRouter()
	rt.Mount(r)

	tokenRec := doJSON(t, r, http.MethodPost, "/api/v1/auth/token",
		map[string]string{"X-API-Key": intAPIKey},
		map[string]any{"telegram_id": intTelegram},
	)
	if tokenRec.Code != http.StatusOK {
		t.Fatalf("token status=%d body=%s", tokenRec.Code, tokenRec.Body.String())
	}
	var tokenOut struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(tokenRec.Body.Bytes(), &tokenOut); err != nil {
		t.Fatal(err)
	}
	authHeader := map[string]string{"Authorization": "Bearer " + tokenOut.AccessToken}

	createRec := doJSON(t, r, http.MethodPost, "/api/v1/projects", authHeader, map[string]any{
		"name": "integration-project",
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create project status=%d body=%s", createRec.Code, createRec.Body.String())
	}
	var project struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}

	archiveRec := doJSON(t, r, http.MethodPost, "/api/v1/projects/"+project.ID+"/archive", authHeader, nil)
	if archiveRec.Code != http.StatusOK {
		t.Fatalf("archive status=%d body=%s", archiveRec.Code, archiveRec.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/api/v1/projects", authHeader, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}
	var listed struct {
		Projects []any `json:"projects"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Projects) != 0 {
		t.Fatalf("archived project should not be listed, got %+v", listed.Projects)
	}
}

func TestAPIIntegrationTaskFlow(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := startPG(t, ctx)
	defer cleanup()

	userRepo := identityinfra.NewRepository(pool)
	user, err := domain.NewUser(intTelegram+1, "Tasks", "UTC", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := userRepo.Upsert(ctx, user); err != nil {
		t.Fatal(err)
	}

	taskRepo := tasksinfra.NewRepository(pool)
	transactor := platformpostgres.NewTransactor(pool)
	eventPub := events.NewPublisher(pool)
	tzReader := identityinfra.NewTimezoneReader(pool)
	createTask := tasksapp.NewCreateTask(taskRepo, eventPub, transactor, projectsinfra.NewProjectReader(pool))
	completeTask := tasksapp.NewCompleteTask(taskRepo, eventPub, transactor)
	listToday := tasksapp.NewListTasksToday(taskRepo, tzReader)

	tokens, err := auth.NewTokenService(intJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	rt := api.NewRouter(api.Deps{
		Log:        slog.Default(),
		APIKey:     intAPIKey,
		Tokens:     tokens,
		GetUser:    identityapp.NewGetUserByTelegram(userRepo),
		CreateTask: createTask,
		Complete:   completeTask,
		ListToday:  listToday,
	})
	r := chi.NewRouter()
	rt.Mount(r)

	tokenRec := doJSON(t, r, http.MethodPost, "/api/v1/auth/token",
		map[string]string{"X-API-Key": intAPIKey},
		map[string]any{"telegram_id": user.TelegramID},
	)
	if tokenRec.Code != http.StatusOK {
		t.Fatalf("token status=%d", tokenRec.Code)
	}
	var tokenOut struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(tokenRec.Body.Bytes(), &tokenOut)
	authHeader := map[string]string{"Authorization": "Bearer " + tokenOut.AccessToken}

	today := time.Now().UTC().Format("2006-01-02")
	createRec := doJSON(t, r, http.MethodPost, "/api/v1/tasks", authHeader, map[string]any{
		"title":    "db task",
		"priority": "medium",
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

	completeRec := doJSON(t, r, http.MethodPost, "/api/v1/tasks/"+created.ID+"/complete", authHeader, nil)
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
		t.Fatalf("status=%q", done.Status)
	}
}

func TestAPIIntegrationProjectsAndReviews(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := startPG(t, ctx)
	defer cleanup()

	userRepo := identityinfra.NewRepository(pool)
	user, err := domain.NewUser(intTelegram+2, "Projects", "UTC", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := userRepo.Upsert(ctx, user); err != nil {
		t.Fatal(err)
	}

	transactor := platformpostgres.NewTransactor(pool)
	eventPub := events.NewPublisher(pool)
	sphereRepo := spheresinfra.NewRepository(pool)
	if err := spheresapp.NewEnsureDefaultSpheres(sphereRepo).Execute(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	spheres, err := spheresapp.NewListSpheres(sphereRepo).Execute(ctx, user.ID)
	if err != nil || len(spheres) == 0 {
		t.Fatal("expected default spheres")
	}
	projectRepo := projectsinfra.NewRepository(pool)
	tzReader := identityinfra.NewTimezoneReader(pool)
	createProject := projectsapp.NewCreateProject(projectRepo, eventPub, transactor)
	listProjects := projectsapp.NewListProjects(projectRepo)
	review := query.NewReview(pool, tzReader, nil)
	priorities := query.NewGetTopPriorities(pool, tzReader)

	tokens, err := auth.NewTokenService(intJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	rt := api.NewRouter(api.Deps{
		Log:           slog.Default(),
		APIKey:        intAPIKey,
		Tokens:        tokens,
		GetUser:       identityapp.NewGetUserByTelegram(userRepo),
		CreateProject: createProject,
		ListProjects:  listProjects,
		Review:        review,
		Priorities:    priorities,
	})
	r := chi.NewRouter()
	rt.Mount(r)

	tokenRec := doJSON(t, r, http.MethodPost, "/api/v1/auth/token",
		map[string]string{"X-API-Key": intAPIKey},
		map[string]any{"telegram_id": user.TelegramID},
	)
	if tokenRec.Code != http.StatusOK {
		t.Fatalf("token status=%d", tokenRec.Code)
	}
	var tokenOut struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(tokenRec.Body.Bytes(), &tokenOut)
	authHeader := map[string]string{"Authorization": "Bearer " + tokenOut.AccessToken}

	createRec := doJSON(t, r, http.MethodPost, "/api/v1/projects", authHeader, map[string]any{
		"name":       "интеграционный проект",
		"sphere_ids": []string{spheres[0].ID.String()},
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create project status=%d body=%s", createRec.Code, createRec.Body.String())
	}

	morningRec := doJSON(t, r, http.MethodGet, "/api/v1/reviews/morning", authHeader, nil)
	if morningRec.Code != http.StatusOK {
		t.Fatalf("morning review status=%d body=%s", morningRec.Code, morningRec.Body.String())
	}
	var morning struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(morningRec.Body.Bytes(), &morning); err != nil || morning.Text == "" {
		t.Fatalf("morning review body=%s", morningRec.Body.String())
	}

	prioRec := doJSON(t, r, http.MethodGet, "/api/v1/priorities", authHeader, nil)
	if prioRec.Code != http.StatusOK {
		t.Fatalf("priorities status=%d body=%s", prioRec.Code, prioRec.Body.String())
	}
}

func startPG(t *testing.T, ctx context.Context) (*pgxpool.Pool, func()) {
	t.Helper()
	if conn := os.Getenv("LIFEOS_DATABASE_URL"); conn != "" {
		return startPGFromURL(t, ctx, conn)
	}
	c, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("lifeos"),
		postgres.WithUsername("lifeos"),
		postgres.WithPassword("lifeos"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	migrationsDir := filepath.Join("..", "..", "..", "migrations")
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

func startPGFromURL(t *testing.T, ctx context.Context, conn string) (*pgxpool.Pool, func()) {
	t.Helper()
	migrationsDir := filepath.Join("..", "..", "..", "migrations")
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
	return pool, func() { pool.Close() }
}
