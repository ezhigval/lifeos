package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai"
)

func TestResolverDegradesOnHTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"busy"}`))
	}))
	t.Cleanup(srv.Close)

	r := NewResolver(NewClient(srv.URL, "test"))
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "сделай что-нибудь"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s, want unknown", got.Type)
	}
}

func TestResolverDegradesOnTimeout(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(500 * time.Millisecond):
		case <-r.Context().Done():
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := NewClientWithTimeout(srv.URL, "test", 30*time.Millisecond)
	r := NewResolver(client)
	start := time.Now()
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "текст"})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s, want unknown", got.Type)
	}
	if elapsed > 300*time.Millisecond {
		t.Fatalf("timeout took too long: %v", elapsed)
	}
}

func TestAssistantDegradesOnDown(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	a := NewAssistant(NewClient(srv.URL, "test"))
	_, err := a.Summarize(context.Background(), ai.SummaryRequest{Tasks: []string{"а"}})
	if err == nil {
		t.Fatal("expected error when ollama down")
	}
}

func TestAssistantSanitizesHTML(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"День с <b>фокусом</b> и <script>x</script>"}}`))
	}))
	t.Cleanup(srv.Close)

	a := NewAssistant(NewClient(srv.URL, "test"))
	got, err := a.Summarize(context.Background(), ai.SummaryRequest{Tasks: []string{"бег"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "<b>фокусом</b>") {
		t.Fatalf("bold lost: %q", got)
	}
	if strings.Contains(got, "<script") {
		t.Fatalf("script leaked: %q", got)
	}
}
