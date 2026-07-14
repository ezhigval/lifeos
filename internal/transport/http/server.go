package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/valentinezhov/lifeos/internal/platform/postgres"
	"github.com/valentinezhov/lifeos/internal/transport/http/api"
)

type Server struct {
	log  *slog.Logger
	addr string
	srv  *http.Server
}

func New(log *slog.Logger, addr string, db *postgres.Pool, traceHTTP bool, apiRouter *api.Router, tgWebhook http.Handler) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	if traceHTTP {
		r.Use(func(next http.Handler) http.Handler {
			return otelhttp.NewHandler(next, "lifeos.http")
		})
	}

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			log.Error("readiness check failed", "error", err)
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	r.Handle("/metrics", promhttp.Handler())

	if apiRouter != nil {
		apiRouter.Mount(r)
		log.Info("rest api enabled", "prefix", "/api/v1")
	}

	if tgWebhook != nil {
		r.Post("/webhook/telegram", tgWebhook.ServeHTTP)
		log.Info("telegram webhook enabled", "path", "/webhook/telegram")
	}

	return &Server{
		log:  log,
		addr: addr,
		srv: &http.Server{
			Addr:              addr,
			Handler:           r,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	s.log.Info("http server listening", "addr", s.addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
