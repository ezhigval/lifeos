package http

import (
	"context"
	"log/slog"
	"mime"
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
	// trusted reverse proxy (Caddy) that sets the leftmost hop correctly.
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
		// Telegram WebView caches HTML aggressively. A stale index points at
		// removed hashed chunks and the module never starts.
		serveMiniAppFile(w, req, filepath.Join(dir, "index.html"), "/index.html")
	}
	// Router-level Compress (and the static test) gzip when no .br/.gz sibling
	// exists. Precompressed files set Content-Encoding so Compress leaves them.
	r.Get("/app", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/app/", http.StatusFound)
	})
	r.Handle("/app/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := strings.TrimPrefix(req.URL.Path, "/app")
		if path == "" || path == "/" {
			serveIndex(w, req)
			return
		}
		// Precompressed siblings are not URLs. A .js.br response with the
		// wrong content type is worse than a miss.
		if strings.HasSuffix(path, ".br") || strings.HasSuffix(path, ".gz") {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.NotFound(w, req)
			return
		}
		full := filepath.Join(dir, filepath.Clean("/"+path))
		root := filepath.Clean(dir)
		if full != root && !strings.HasPrefix(full, root+string(filepath.Separator)) {
			http.NotFound(w, req)
			return
		}
		if st, err := os.Stat(full); err == nil && !st.IsDir() {
			serveMiniAppFile(w, req, full, path)
			return
		}
		// Missing JS/CSS must 404. Falling back to index.html returns
		// text/html for a module URL, and the WebView never executes it.
		if isMiniAppAssetPath(path) {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.NotFound(w, req)
			return
		}
		// SPA fallback for client-side routes (BrowserRouter basename=/app).
		serveIndex(w, req)
	}))
}

func serveMiniAppFile(w http.ResponseWriter, req *http.Request, full, urlPath string) {
	// Telegram Desktop fetches scripts with CORS from a webview origin that
	// is not the page. Without this header the script is dropped.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", miniAppCacheControl(urlPath))
	w.Header().Set("X-Content-Type-Options", "nosniff")

	payload := full
	if variant, encoding, ok := precompressed(full, req.Header.Get("Accept-Encoding")); ok {
		payload = variant
		w.Header().Set("Content-Encoding", encoding)
		w.Header().Add("Vary", "Accept-Encoding")
		// A range of brotli bytes is not a valid partial script.
		req = req.Clone(req.Context())
		req.Header.Del("Range")
	}
	if ctype := mime.TypeByExtension(filepath.Ext(full)); ctype != "" {
		w.Header().Set("Content-Type", ctype)
	}

	f, err := os.Open(payload)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		http.NotFound(w, req)
		return
	}
	http.ServeContent(w, req, filepath.Base(full), st.ModTime(), f)
}

// precompressed picks a build-time .br or .gz sibling. chi's Compress middleware
// leaves the body alone once Content-Encoding is set, so this is not wrapped twice.
func precompressed(full, accept string) (variant, encoding string, ok bool) {
	if acceptToken(accept, "br") {
		if st, err := os.Stat(full + ".br"); err == nil && !st.IsDir() {
			return full + ".br", "br", true
		}
	}
	if acceptToken(accept, "gzip") {
		if st, err := os.Stat(full + ".gz"); err == nil && !st.IsDir() {
			return full + ".gz", "gzip", true
		}
	}
	return "", "", false
}

func acceptToken(header, coding string) bool {
	for _, part := range strings.Split(header, ",") {
		name, _, _ := strings.Cut(strings.TrimSpace(part), ";")
		name = strings.TrimSpace(name)
		if name == coding || name == "*" {
			return true
		}
	}
	return false
}

func miniAppCacheControl(urlPath string) string {
	base := strings.ToLower(filepath.Base(urlPath))
	// HTML and the service worker must revalidate. A cached sw.js would keep
	// serving an old asset policy after deploy.
	if base == "index.html" || base == "sw.js" {
		return "no-store"
	}
	// Vite fingerprints files under /assets/. A hashed telegram-web-app.<hash>.js
	// can be immutable too. The unhashed SDK stays revalidated: the page injects
	// /app/telegram-web-app.js after React mounts.
	if strings.HasPrefix(urlPath, "/assets/") || fingerprintedSDK(base) {
		return "public, max-age=31536000, immutable"
	}
	return "public, max-age=86400"
}

func fingerprintedSDK(base string) bool {
	const prefix = "telegram-web-app."
	if !strings.HasPrefix(base, prefix) || !strings.HasSuffix(base, ".js") {
		return false
	}
	hash := strings.TrimSuffix(strings.TrimPrefix(base, prefix), ".js")
	if len(hash) < 8 {
		return false
	}
	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func isMiniAppAssetPath(urlPath string) bool {
	switch strings.ToLower(filepath.Ext(urlPath)) {
	case ".js", ".mjs", ".css", ".map", ".json", ".svg", ".png", ".jpg", ".jpeg", ".webp", ".gif", ".ico", ".woff", ".woff2", ".ttf", ".txt", ".webmanifest", ".wasm", ".html":
		return true
	default:
		return false
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
