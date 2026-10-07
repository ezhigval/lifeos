package http

import (
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
	if cc := rr.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Fatalf("index cache %q", cc)
	}
}
