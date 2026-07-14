package http

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsRouteRegisters(t *testing.T) {
	t.Parallel()

	s := New(slog.Default(), ":0", nil, false, nil, nil, Options{})
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if body == "" {
		t.Fatal("expected prometheus metrics body")
	}
}

func TestHealthRouteRegisters(t *testing.T) {
	t.Parallel()

	s := New(slog.Default(), ":0", nil, false, nil, nil, Options{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Fatalf("health status=%d body=%q", rec.Code, rec.Body.String())
	}
}
