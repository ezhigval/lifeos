package vision_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai/vision"
)

func TestImageToUserText(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Fatal("missing auth")
		}
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatal(err)
		}
		if body["model"] != "meta-llama/llama-4-scout-17b-16e-instruct" {
			t.Fatalf("model=%v", body["model"])
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"запиши расход 320 еда"}}]}`))
	}))
	defer srv.Close()

	c := vision.NewClient(srv.URL, "key", "")
	got, err := c.ImageToUserText(context.Background(), []byte("fake-jpeg"), "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if got != "запиши расход 320 еда" {
		t.Fatalf("got %q", got)
	}
}

func TestImageToUserTextRetries5xx(t *testing.T) {
	t.Parallel()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`bad gateway`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"добавь задачу купить молоко"}}]}`))
	}))
	defer srv.Close()

	c := vision.NewClient(srv.URL, "key", "m")
	got, err := c.ImageToUserText(context.Background(), []byte("x"), "image/png")
	if err != nil {
		t.Fatal(err)
	}
	if got != "добавь задачу купить молоко" {
		t.Fatalf("got %q", got)
	}
	if n.Load() != 3 {
		t.Fatalf("attempts=%d", n.Load())
	}
}

func TestImageToUserTextRequiresKey(t *testing.T) {
	t.Parallel()
	c := vision.NewClient("", "", "")
	_, err := c.ImageToUserText(context.Background(), []byte("x"), "image/jpeg")
	if err == nil {
		t.Fatal("expected error")
	}
}
