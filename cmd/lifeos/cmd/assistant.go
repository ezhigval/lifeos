package cmd

import (
	"log/slog"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/composite"
	"github.com/valentinezhov/lifeos/internal/ai/ollama"
	"github.com/valentinezhov/lifeos/internal/ai/openaiapi"
	"github.com/valentinezhov/lifeos/internal/ai/templateassistant"
	"github.com/valentinezhov/lifeos/internal/platform/config"
)

// newAssistant wires review summarization:
//  1. when LIFEOS_LLM_ENABLED — try configured LLM (sanitized HTML, only <b> allowed)
//  2. always fall back to templateassistant on error/timeout/empty/unsafe output
//
// Default compose does not start Ollama; LLM stays off unless enabled in env.
func newAssistant(cfg config.Config, log *slog.Logger) ai.Assistant {
	template := templateassistant.New()
	if !cfg.LLMEnabled {
		return template
	}
	llm := newLLMAssistant(cfg, log)
	return composite.NewFallbackAssistant(llm, template)
}

func newLLMAssistant(cfg config.Config, log *slog.Logger) ai.Assistant {
	if cfg.LLMProvider == "ollama" {
		client := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
		log.Info("llm assistant enabled",
			"provider", "ollama",
			"base_url", cfg.OllamaURL,
			"model", cfg.OllamaModel,
		)
		return ollama.NewAssistant(client)
	}
	client := openaiapi.NewClient(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	log.Info("llm assistant enabled",
		"provider", "openai",
		"base_url", cfg.LLMBaseURL,
		"model", cfg.LLMModel,
	)
	return openaiapi.NewAssistant(client)
}
