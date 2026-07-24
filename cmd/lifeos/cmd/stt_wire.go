package cmd

import (
	"log/slog"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/stt"
	"github.com/valentinezhov/lifeos/internal/platform/config"
)

func newSpeechToText(cfg config.Config, log *slog.Logger) ai.SpeechToText {
	if !cfg.STTEnabled {
		return nil
	}
	if cfg.STTAPIKey == "" {
		log.Warn("stt enabled but no API key (LIFEOS_STT_API_KEY or LIFEOS_LLM_API_KEY) — voice/video notes disabled")
		return nil
	}
	log.Info("speech-to-text enabled",
		"base_url", cfg.STTBaseURL,
		"model", cfg.STTModel,
	)
	return stt.NewClient(cfg.STTBaseURL, cfg.STTAPIKey, cfg.STTModel)
}
