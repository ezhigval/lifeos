package cmd

import (
	"log/slog"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/composite"
	"github.com/valentinezhov/lifeos/internal/ai/ollama"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
	"github.com/valentinezhov/lifeos/internal/platform/config"
)

func newIntentResolver(cfg config.Config, log *slog.Logger) ai.IntentResolver {
	primary := rulebased.NewResolver()
	if !cfg.LLMEnabled {
		return primary
	}
	client := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
	llm := ollama.NewResolver(client)
	log.Info("llm intent resolver enabled", "ollama_url", cfg.OllamaURL, "model", cfg.OllamaModel)
	return composite.NewFallbackResolver(primary, llm, func() bool { return true })
}
