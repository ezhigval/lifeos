package http

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func TestMiniAppAssetsAreCachedAndGzipped(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>index</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Repeat("function lifeos(){return 1}\n", 200)
	if err := os.WriteFile(filepath.Join(dir, "assets", "index-abc.js"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	mountMiniApp(r, dir)

	req := httptest.NewRequest(http.MethodGet, "/app/assets/index-abc.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if cc := rr.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("cache %q", cc)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("asset acao %q", got)
	}
	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("encoding %q len %d", rr.Header().Get("Content-Encoding"), rr.Body.Len())
	}
	if rr.Body.Len() >= len(body) {
		t.Fatalf("gzip did not shrink: %d >= %d", rr.Body.Len(), len(body))
	}

	req = httptest.NewRequest(http.MethodGet, "/app/", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("index status %d", rr.Code)
	}
	if cc := rr.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("index cache %q", cc)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("index acao %q", got)
	}
}

func newMiniAppServer(t *testing.T) (*Server, string) {
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
	return New(slog.Default(), ":0", nil, false, nil, nil, Options{StaticDir: dir}), dir
}

func TestMiniAppIndexIsNotCached(t *testing.T) {
	t.Parallel()
	s, _ := newMiniAppServer(t)

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
	s, _ := newMiniAppServer(t)

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
	s, _ := newMiniAppServer(t)

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
	s, _ := newMiniAppServer(t)

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
	s, _ := newMiniAppServer(t)

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

func TestMiniAppPrecompressedPrefersBrotliAndDoesNotWrapGzip(t *testing.T) {
	t.Parallel()
	s, dir := newMiniAppServer(t)

	raw := []byte("export const booted = true\n")
	if err := os.WriteFile(filepath.Join(dir, "assets", "index-abc.js.br"), []byte("BROTLI"), 0o644); err != nil {
		t.Fatal(err)
	}
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	if _, err := zw.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "index-abc.js.gz"), gz.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	brReq := httptest.NewRequest(http.MethodGet, "/app/assets/index-abc.js", nil)
	brReq.Header.Set("Accept-Encoding", "gzip, deflate, br")
	brRec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(brRec, brReq)
	if brRec.Code != http.StatusOK {
		t.Fatalf("br status = %d", brRec.Code)
	}
	if brRec.Header().Get("Content-Encoding") != "br" {
		t.Fatalf("Content-Encoding = %q, want br", brRec.Header().Get("Content-Encoding"))
	}
	if brRec.Body.String() != "BROTLI" {
		t.Fatalf("body = %q", brRec.Body.String())
	}
	if !strings.Contains(brRec.Header().Get("Content-Type"), "javascript") {
		t.Fatalf("Content-Type = %q", brRec.Header().Get("Content-Type"))
	}
	if !strings.Contains(brRec.Header().Get("Vary"), "Accept-Encoding") {
		t.Fatalf("Vary = %q", brRec.Header().Get("Vary"))
	}

	gzReq := httptest.NewRequest(http.MethodGet, "/app/assets/index-abc.js", nil)
	gzReq.Header.Set("Accept-Encoding", "gzip")
	gzRec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(gzRec, gzReq)
	if gzRec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", gzRec.Header().Get("Content-Encoding"))
	}
	zr, err := gzip.NewReader(gzRec.Body)
	if err != nil {
		t.Fatalf("body is not a single gzip stream: %v", err)
	}
	got, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatalf("gunzip = %q, want %q", got, raw)
	}
}

func TestMiniAppFingerprintedSDKIsImmutable(t *testing.T) {
	t.Parallel()
	s, dir := newMiniAppServer(t)
	name := "telegram-web-app.abcdef1234.js"
	if err := os.WriteFile(filepath.Join(dir, name), []byte("/* sdk */\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app/"+name, nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", cc)
	}
}

func TestMiniAppServiceWorkerIsNotStored(t *testing.T) {
	t.Parallel()
	s, dir := newMiniAppServer(t)
	if err := os.WriteFile(filepath.Join(dir, "sw.js"), []byte("/* sw */\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app/sw.js", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cc)
	}
}

func TestMiniAppRejectsDirectCompressedURL(t *testing.T) {
	t.Parallel()
	s, _ := newMiniAppServer(t)

	req := httptest.NewRequest(http.MethodGet, "/app/assets/index-abc.js.br", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "export const") || strings.Contains(rec.Body.String(), "LifeOS") {
		t.Fatalf("compressed sibling leaked: %q", rec.Body.String())
	}
}

func TestMiniAppUnhashedPublicFileIsRevalidated(t *testing.T) {
	t.Parallel()
	s, _ := newMiniAppServer(t)

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
