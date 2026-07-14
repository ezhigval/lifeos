package composite

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type stubAssistant struct {
	text  string
	err   error
	delay time.Duration
}

func (s stubAssistant) Summarize(ctx context.Context, _ ai.SummaryRequest) (string, error) {
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return s.text, s.err
}

func TestFallbackAssistantUsesPrimary(t *testing.T) {
	t.Parallel()
	a := NewFallbackAssistant(
		stubAssistant{text: "LLM summary"},
		stubAssistant{text: "template"},
	)
	got, err := a.Summarize(context.Background(), ai.SummaryRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "LLM summary" {
		t.Fatalf("got %q", got)
	}
}

func TestFallbackAssistantUsesFallbackOnError(t *testing.T) {
	t.Parallel()
	a := NewFallbackAssistant(
		stubAssistant{err: errors.New("down")},
		stubAssistant{text: "template ok"},
	)
	got, err := a.Summarize(context.Background(), ai.SummaryRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "template ok" {
		t.Fatalf("got %q", got)
	}
}

func TestFallbackAssistantUsesFallbackOnEmpty(t *testing.T) {
	t.Parallel()
	a := NewFallbackAssistant(
		stubAssistant{text: "   "},
		stubAssistant{text: "template empty"},
	)
	got, err := a.Summarize(context.Background(), ai.SummaryRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "template empty" {
		t.Fatalf("got %q", got)
	}
}

func TestFallbackAssistantUsesFallbackOnTimeout(t *testing.T) {
	t.Parallel()
	a := NewFallbackAssistant(
		stubAssistant{text: "late llm", delay: 200 * time.Millisecond},
		stubAssistant{text: "template timeout"},
	).WithTimeout(20 * time.Millisecond)

	start := time.Now()
	got, err := a.Summarize(context.Background(), ai.SummaryRequest{})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if got != "template timeout" {
		t.Fatalf("got %q", got)
	}
	if elapsed > 150*time.Millisecond {
		t.Fatalf("timeout took too long: %v", elapsed)
	}
}

func TestFallbackAssistantSanitizesUnsafeHTML(t *testing.T) {
	t.Parallel()
	a := NewFallbackAssistant(
		stubAssistant{text: `Focus <b>here</b> <script>x</script>`},
		stubAssistant{text: "template"},
	)
	got, err := a.Summarize(context.Background(), ai.SummaryRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "<b>here</b>") {
		t.Fatalf("bold lost: %q", got)
	}
	if strings.Contains(got, "<script") {
		t.Fatalf("script leaked: %q", got)
	}
	if !strings.Contains(got, "&lt;script") {
		t.Fatalf("expected escaped script: %q", got)
	}
}

func TestFallbackAssistantReturnsTemplateAsIs(t *testing.T) {
	t.Parallel()
	a := NewFallbackAssistant(
		stubAssistant{err: errors.New("down")},
		stubAssistant{text: `template &lt;safe&gt;`},
	)
	got, err := a.Summarize(context.Background(), ai.SummaryRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got != `template &lt;safe&gt;` {
		t.Fatalf("got %q", got)
	}
}
