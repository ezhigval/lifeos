package telegram_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valentinezhov/lifeos/internal/transport/telegram"
)

type stubHandler struct {
	called bool
	err    error
}

func (s *stubHandler) HandleUpdate(_ context.Context, _ telegram.Update) error {
	s.called = true
	return s.err
}

func TestWebhookRejectsInvalidSecret(t *testing.T) {
	t.Parallel()
	h := &stubHandler{}
	wh := telegram.NewWebhook(h, "secret", slog.Default())

	body := []byte(`{"update_id":1}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/telegram", bytes.NewReader(body))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong")
	rec := httptest.NewRecorder()
	wh.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if h.called {
		t.Fatal("handler should not be called")
	}
}

func TestWebhookAcceptsValidUpdate(t *testing.T) {
	t.Parallel()
	h := &stubHandler{}
	wh := telegram.NewWebhook(h, "secret", slog.Default())

	body := []byte(`{"update_id":42}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/telegram", bytes.NewReader(body))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "secret")
	rec := httptest.NewRecorder()
	wh.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !h.called {
		t.Fatal("handler should be called")
	}
}

func TestWebhookReturns500WhenHandlerFails(t *testing.T) {
	t.Parallel()
	h := &stubHandler{err: errors.New("db down")}
	wh := telegram.NewWebhook(h, "secret", slog.Default())

	body := []byte(`{"update_id":7}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook/telegram", bytes.NewReader(body))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "secret")
	rec := httptest.NewRecorder()
	wh.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
