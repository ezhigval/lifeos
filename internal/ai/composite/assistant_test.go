package composite

import (
	"context"
	"errors"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type stubAssistant struct {
	text string
	err  error
}

func (s stubAssistant) Summarize(context.Context, ai.SummaryRequest) (string, error) {
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
