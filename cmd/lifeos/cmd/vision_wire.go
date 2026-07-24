package cmd

import (
	"log/slog"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/vision"
	"github.com/valentinezhov/lifeos/internal/platform/config"
)

func newVision(cfg config.Config, log *slog.Logger) ai.Vision {
	if !cfg.VisionEnabled {
		return nil
	}
	if cfg.VisionAPIKey == "" {
		log.Warn("vision enabled but no API key (LIFEOS_VISION_API_KEY or LIFEOS_LLM_API_KEY) — photo understanding disabled")
		return nil
	}
	log.Info("vision enabled",
		"base_url", cfg.VisionBaseURL,
		"model", cfg.VisionModel,
	)
	return vision.NewClient(cfg.VisionBaseURL, cfg.VisionAPIKey, cfg.VisionModel)
}
