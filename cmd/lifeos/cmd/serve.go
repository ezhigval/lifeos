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
	if err := validateRuntimeConfig(cfg); err != nil {
		return err
	}
	warnSoftConfig(cfg)

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

	if rt.tgClient != nil {
		// Advertise "/" commands (incl. /triage) in the Telegram client menu.
		if err := rt.tgClient.SetMyCommands(ctx, tg.DefaultBotCommands()); err != nil {
			log.Warn("set bot commands failed", "error", err)
		}
	}

	if rt.tgClient != nil && cfg.TelegramMode == "webhook" {
		// Registration is done externally (lifeos telegram set-webhook) from a
		// network where api.telegram.org is reachable — e.g. when the host
		// blocks outbound Telegram (Yandex Cloud). Here we only verify that the
		// webhook Telegram actually points at matches our config; on mismatch
		// we attempt to (re)register once but never abort startup because of it.
		if info, err := rt.tgClient.GetWebhookInfo(ctx); err != nil {
			log.Warn("cannot query telegram webhook info (host may block api.telegram.org)", "error", err)
		} else if info.URL != cfg.TelegramWebhookURL {
			log.Warn("telegram webhook mismatch, attempting registration", "current", info.URL, "want", cfg.TelegramWebhookURL)
			if err := tg.RegisterWebhook(ctx, rt.tgClient, cfg.TelegramWebhookURL, cfg.TelegramWebhookSecret); err != nil {
				log.Error("telegram webhook registration failed; run `lifeos telegram set-webhook` from a reachable network", "error", err)
			} else {
				log.Info("telegram webhook registered", "url", cfg.TelegramWebhookURL)
			}
		} else {
			log.Info("telegram webhook active", "url", info.URL, "pending", info.PendingUpdateCount)
		}
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
