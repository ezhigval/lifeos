package ollama_test

import (
	"context"
	"os"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/composite"
	"github.com/valentinezhov/lifeos/internal/ai/ollama"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
)

// Live smoke against local Ollama (or mock) when LIFEOS_OLLAMA_SMOKE=1.
func TestLiveOllamaFallbackSmoke(t *testing.T) {
	if os.Getenv("LIFEOS_OLLAMA_SMOKE") != "1" {
		t.Skip("set LIFEOS_OLLAMA_SMOKE=1 to run")
	}
	base := os.Getenv("LIFEOS_OLLAMA_URL")
	if base == "" {
		base = "http://127.0.0.1:11434"
	}
	model := os.Getenv("LIFEOS_OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2:1b"
	}
	primary := rulebased.NewResolver()
	llm := ollama.NewResolver(ollama.NewClient(base, model))
	r := composite.NewFallbackResolver(primary, llm, func() bool { return true })

	got, err := r.Resolve(context.Background(), ai.ResolveInput{
		Text: "задачи на сегодня", Language: "ru",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentTaskListToday {
		t.Fatalf("rulebased path: got %s", got.Type)
	}

	got, err = r.Resolve(context.Background(), ai.ResolveInput{
		Text: "хмм разбери пожалуйста входящие письма", Language: "ru",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type == ai.IntentUnknown {
		t.Fatalf("llm fallback should classify unknown phrase, got unknown")
	}
	t.Logf("llm fallback: type=%s title=%q msg=%q conf=%.2f", got.Type, got.Title, got.Message, got.Confidence)
}
