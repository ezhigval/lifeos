package http

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newMiniAppServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	index := "<!doctype html><title>LifeOS</title><script type=\"module\" src=\"/app/assets/index-abc.js\"></script>"
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	js := "export const booted = true\n"
	if err := os.WriteFile(filepath.Join(dir, "assets", "index-abc.js"), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "telegram-web-app.js"), []byte("/* sdk */\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return New(slog.Default(), ":0", nil, false, nil, nil, Options{StaticDir: dir})
}

func TestMiniAppIndexIsNotCached(t *testing.T) {
	t.Parallel()
	s := newMiniAppServer(t)

	req := httptest.NewRequest(http.MethodGet, "/app/", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "LifeOS") {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cc)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("missing nosniff: %v", rec.Header())
	}
}

func TestMiniAppRedirectsBarePath(t *testing.T) {
	t.Parallel()
	s := newMiniAppServer(t)

	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/app/" {
		t.Fatalf("Location = %q", loc)
	}
}

func TestMiniAppHashedAssetIsImmutableJS(t *testing.T) {
	t.Parallel()
	s := newMiniAppServer(t)

	req := httptest.NewRequest(http.MethodGet, "/app/assets/index-abc.js", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "javascript") {
		t.Fatalf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}
	if strings.Contains(rec.Body.String(), "<!doctype") {
		t.Fatal("asset response is HTML")
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", cc)
	}
}

func TestMiniAppMissingScriptDoesNotFallBackToIndex(t *testing.T) {
	t.Parallel()
	s := newMiniAppServer(t)

	req := httptest.NewRequest(http.MethodGet, "/app/assets/index-stale.js", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "LifeOS") || strings.Contains(strings.ToLower(rec.Header().Get("Content-Type")), "html") {
		t.Fatalf("missing script fell back to HTML: type=%q body=%q", rec.Header().Get("Content-Type"), rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cc)
	}
}

func TestMiniAppSPAFallbackForClientRoutes(t *testing.T) {
	t.Parallel()
	s := newMiniAppServer(t)

	req := httptest.NewRequest(http.MethodGet, "/app/more/habits", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "LifeOS") {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cc)
	}
}

func TestMiniAppUnhashedPublicFileIsRevalidated(t *testing.T) {
	t.Parallel()
	s := newMiniAppServer(t)

	req := httptest.NewRequest(http.MethodGet, "/app/telegram-web-app.js?v=20261007", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "sdk") {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=86400" {
		t.Fatalf("Cache-Control = %q", cc)
	}
}
