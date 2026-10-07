package http

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/valentinezhov/lifeos/internal/platform/postgres"
	"github.com/valentinezhov/lifeos/internal/transport/http/api"
)

type Options struct {
	StaticDir string
}

type Server struct {
	log  *slog.Logger
	addr string
	srv  *http.Server
}

func New(log *slog.Logger, addr string, db *postgres.Pool, traceHTTP bool, apiRouter *api.Router, tgWebhook http.Handler, opts Options) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	// Mini App JS is ~400KB. Compress it before it crosses the tunnel.
	r.Use(middleware.Compress(5))
	// RealIP is deprecated (spoofable X-Forwarded-For) but LifeOS sits behind a
	// trusted reverse proxy (Caddy/Fly) that sets the leftmost hop correctly.
	r.Use(middleware.RealIP) //nolint:staticcheck // SA1019: trusted proxy edge
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

	if dir := strings.TrimSpace(opts.StaticDir); dir != "" {
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			mountMiniApp(r, dir)
			log.Info("mini app static enabled", "path", "/app/", "dir", dir)
		} else {
			log.Warn("mini app static dir missing", "dir", dir)
		}
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

func mountMiniApp(r chi.Router, dir string) {
	serveIndex := func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(dir, "index.html"))
	}
	r.Get("/app", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/app/", http.StatusFound)
	})
	r.Handle("/app/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := strings.TrimPrefix(req.URL.Path, "/app")
		if path == "" || path == "/" {
			setMiniAppCache(w, "/")
			serveIndex(w, req)
			return
		}
		full := filepath.Join(dir, filepath.Clean("/"+path))
		root := filepath.Clean(dir)
		if full != root && !strings.HasPrefix(full, root+string(filepath.Separator)) {
			http.NotFound(w, req)
			return
		}
		if st, err := os.Stat(full); err == nil && !st.IsDir() {
			setMiniAppCache(w, path)
			http.ServeFile(w, req, full)
			return
		}
		// SPA fallback for client-side routes (BrowserRouter basename=/app).
		setMiniAppCache(w, "/")
		serveIndex(w, req)
	}))
}

func setMiniAppCache(w http.ResponseWriter, path string) {
	// Telegram Desktop fetches type=module with CORS from a webview origin
	// that is not the page. Without this header the module is dropped and
	// the static boot line stays on screen.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch {
	case strings.HasPrefix(path, "/assets/"):
		// Filenames are content-hashed by Vite.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	case path == "/telegram-web-app.js":
		w.Header().Set("Cache-Control", "public, max-age=86400")
	default:
		w.Header().Set("Cache-Control", "no-cache")
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
