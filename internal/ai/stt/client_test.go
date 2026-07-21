package stt_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai/stt"
)

func TestClientTranscribe(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio/transcriptions" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Fatal("missing auth")
		}
		if _, err := io.ReadAll(r.Body); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"text":"купи хлеб"}`))
	}))
	defer srv.Close()

	c := stt.NewClient(srv.URL, "key", "whisper-large-v3-turbo")
	got, err := c.Transcribe(context.Background(), []byte("fake-ogg"), "voice.ogg", "ru")
	if err != nil {
		t.Fatal(err)
	}
	if got != "купи хлеб" {
		t.Fatalf("got %q", got)
	}
}

func TestClientRequiresKey(t *testing.T) {
	t.Parallel()
	c := stt.NewClient("", "", "")
	_, err := c.Transcribe(context.Background(), []byte("x"), "a.ogg", "ru")
	if err == nil {
		t.Fatal("expected error")
	}
}
