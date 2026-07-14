package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	identityinfra "github.com/valentinezhov/lifeos/internal/identity/infra"
	"github.com/valentinezhov/lifeos/internal/platform/config"
	"github.com/valentinezhov/lifeos/internal/platform/logging"
	platformotel "github.com/valentinezhov/lifeos/internal/platform/otel"
	"github.com/valentinezhov/lifeos/internal/platform/postgres"
	settingsapp "github.com/valentinezhov/lifeos/internal/settings/app"
	settingsinfra "github.com/valentinezhov/lifeos/internal/settings/infra"
	httptransport "github.com/valentinezhov/lifeos/internal/transport/http"
	tg "github.com/valentinezhov/lifeos/internal/transport/telegram"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start LifeOS server",
	Run: func(_ *cobra.Command, _ []string) {
		exitOnError(runServe())
	},
}

func runServe() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logging.New(cfg.LogLevel, cfg.LogFormat)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownOtel, err := platformotel.Init(ctx, platformotel.Config{
		Enabled:     cfg.OtelEnabled,
		Endpoint:    cfg.OtelEndpoint,
		ServiceName: "lifeos",
	})
	if err != nil {
		return fmt.Errorf("init otel: %w", err)
	}
	defer func() {
		if err := shutdownOtel(context.Background()); err != nil {
			log.Error("shutdown otel", "error", err)
		}
	}()

	pool, err := postgres.New(ctx, cfg.DatabaseURL, cfg.OtelEnabled)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := seedIfConfigured(ctx, log, cfg, pool); err != nil {
		return err
	}

	rt, err := newRuntime(ctx, cfg, log, pool)
	if err != nil {
		return err
	}

	apiRouter, err := rt.apiRouter(cfg, log)
	if err != nil {
		return err
	}

	var tgWebhook http.Handler
	if rt.handler != nil && cfg.TelegramMode == "webhook" {
		tgWebhook = tg.NewWebhook(rt.handler, cfg.TelegramWebhookSecret, log)
	}

	httpSrv := httptransport.New(log, cfg.HTTPAddr, pool, cfg.OtelEnabled, apiRouter, tgWebhook, httptransport.Options{
		StaticDir: cfg.StaticDir,
	})

	if rt.tgClient != nil && cfg.MiniAppURL != "" {
		if err := rt.tgClient.SetChatMenuButton(ctx, "Mini App", cfg.MiniAppURL); err != nil {
			log.Warn("set chat menu button failed", "error", err, "url", cfg.MiniAppURL)
		} else {
			log.Info("telegram mini app menu button set", "url", cfg.MiniAppURL)
		}
	}

	if rt.tgClient != nil && cfg.TelegramMode == "webhook" {
		if err := tg.RegisterWebhook(ctx, rt.tgClient, cfg.TelegramWebhookURL, cfg.TelegramWebhookSecret); err != nil {
			return fmt.Errorf("register telegram webhook: %w", err)
		}
		log.Info("telegram webhook registered", "url", cfg.TelegramWebhookURL)
		defer func() {
			clearCtx, clearCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer clearCancel()
			if err := tg.ClearWebhook(clearCtx, rt.tgClient); err != nil {
				log.Warn("clear telegram webhook failed", "error", err)
			}
		}()
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, gctx := errgroup.WithContext(sigCtx)

	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		log.Info("shutting down http server")
		return httpSrv.Shutdown(shutdownCtx)
	})

	g.Go(func() error {
		return httpSrv.Start()
	})

	if rt.poller != nil && cfg.TelegramMode == "polling" {
		g.Go(func() error {
			return rt.poller.Run(gctx)
		})
	}

	if rt.sched != nil {
		g.Go(func() error {
			return rt.sched.Run(gctx)
		})
	}

	log.Info("lifeos started", "http", cfg.HTTPAddr, "telegram_mode", cfg.TelegramMode)
	return g.Wait()
}

func seedIfConfigured(ctx context.Context, log interface {
	Info(string, ...any)
}, cfg config.Config, pool *postgres.Pool) error {
	if cfg.SeedTelegramID <= 0 {
		return nil
	}

	userRepo := identityinfra.NewRepository(pool.Pool)
	settingsRepo := settingsinfra.NewRepository(pool.Pool)

	seed := identityapp.NewSeedUser(userRepo)
	user, err := seed.Execute(ctx, identityapp.SeedInput{
		TelegramID:  cfg.SeedTelegramID,
		DisplayName: cfg.SeedDisplayName,
		Timezone:    cfg.SeedTimezone,
	})
	if err != nil {
		return fmt.Errorf("seed user: %w", err)
	}

	ensureSettings := settingsapp.NewEnsureDefaults(settingsRepo)
	if err := ensureSettings.Execute(ctx, user.ID); err != nil {
		return fmt.Errorf("seed settings: %w", err)
	}

	log.Info("seed user ensured", "user_id", user.ID.String(), "telegram_id", user.TelegramID)
	return nil
}
