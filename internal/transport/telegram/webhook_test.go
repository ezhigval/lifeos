package telegram_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valentinezhov/lifeos/internal/transport/telegram"
)

type stubHandler struct {
	called bool
}

func (s *stubHandler) HandleUpdate(_ context.Context, _ telegram.Update) error {
	s.called = true
	return nil
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
