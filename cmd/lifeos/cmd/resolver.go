package cmd

import (
	"log/slog"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/composite"
	"github.com/valentinezhov/lifeos/internal/ai/ollama"
	"github.com/valentinezhov/lifeos/internal/ai/openaiapi"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
	"github.com/valentinezhov/lifeos/internal/platform/config"
)

// newIntentResolver wires production resolve order:
//  1. rule-based (always) for known intents — fast, offline
//  2. optional LLM only when LIFEOS_LLM_ENABLED and primary returns unknown
//
// LLM down/timeouts degrade to unknown without failing the request.
func newIntentResolver(cfg config.Config, log *slog.Logger) ai.IntentResolver {
	primary := rulebased.NewResolver()
	if !cfg.LLMEnabled {
		return primary
	}
	llm := newLLMResolver(cfg, log)
	return composite.NewFallbackResolver(primary, llm, func() bool { return true })
}

func newLLMResolver(cfg config.Config, log *slog.Logger) ai.IntentResolver {
	if cfg.LLMProvider == "ollama" {
		client := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
		log.Info("llm intent resolver enabled",
			"provider", "ollama",
			"base_url", cfg.OllamaURL,
			"model", cfg.OllamaModel,
		)
		return ollama.NewResolver(client)
	}
	client := openaiapi.NewClient(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	log.Info("llm intent resolver enabled",
		"provider", "openai",
		"base_url", cfg.LLMBaseURL,
		"model", cfg.LLMModel,
	)
	return openaiapi.NewResolver(client)
}
