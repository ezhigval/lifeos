package cmd

import (
	"log/slog"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/composite"
	"github.com/valentinezhov/lifeos/internal/ai/ollama"
	"github.com/valentinezhov/lifeos/internal/ai/templateassistant"
	"github.com/valentinezhov/lifeos/internal/platform/config"
)

// newAssistant wires review summarization:
//  1. when LIFEOS_LLM_ENABLED — try Ollama (sanitized HTML, only <b> allowed)
//  2. always fall back to templateassistant on error/timeout/empty/unsafe output
//
// Default compose does not start Ollama; LLM stays off unless enabled in env.
func newAssistant(cfg config.Config, log *slog.Logger) ai.Assistant {
	template := templateassistant.New()
	if !cfg.LLMEnabled {
		return template
	}
	client := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
	llm := ollama.NewAssistant(client)
	log.Info("llm assistant enabled", "ollama_url", cfg.OllamaURL, "model", cfg.OllamaModel)
	return composite.NewFallbackAssistant(llm, template)
}
