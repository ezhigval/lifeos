package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/config"
)

func validateRuntimeConfig(cfg config.Config) error {
	var errs []string

	if strings.TrimSpace(cfg.TelegramBotToken) == "" && !envTruthy("LIFEOS_ALLOW_NO_TELEGRAM") {
		errs = append(errs, "TELEGRAM_BOT_TOKEN is required for serve (bot polling + notifications); set LIFEOS_ALLOW_NO_TELEGRAM=true only for API-only local runs")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" && !envTruthy("LIFEOS_ALLOW_NO_API") {
		errs = append(errs, "LIFEOS_JWT_SECRET is required for serve (Mini App / REST API); set LIFEOS_ALLOW_NO_API=true only for bot-only local runs")
	}
	if cfg.LLMEnabled {
		switch strings.ToLower(strings.TrimSpace(cfg.LLMProvider)) {
		case "openai", "openai-compatible", "openai_compatible", "compatible":
			if strings.TrimSpace(cfg.LLMAPIKey) == "" {
				errs = append(errs, "LIFEOS_LLM_API_KEY (or LIFEOS_OPENAI_API_KEY) is required when LIFEOS_LLM_PROVIDER=openai")
			}
		case "ollama", "":
			if err := rejectMockOllamaUnlessAllowed(cfg); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}
	if cfg.LLMAgentEnabled && !cfg.LLMEnabled {
		errs = append(errs, "LIFEOS_LLM_AGENT_ENABLED=true requires LIFEOS_LLM_ENABLED=true")
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
}

func warnSoftConfig(cfg config.Config) {
	if strings.TrimSpace(cfg.LearningSalt) == "" ||
		cfg.LearningSalt == "lifeos-dev-learning-salt" ||
		cfg.LearningSalt == "lifeos-dev-learning-salt-change-me" {
		fmt.Fprintln(os.Stderr, "warning: set a unique LIFEOS_LEARNING_SALT in production (anonymous learning hash)")
	}
	if strings.TrimSpace(cfg.MiniAppURL) == "" && strings.TrimSpace(cfg.StaticDir) != "" {
		fmt.Fprintln(os.Stderr, "warning: LIFEOS_MINIAPP_URL is empty while static Mini App is mounted — bot deep-links will be incomplete")
	}
	if envTruthy("LIFEOS_ALLOW_NO_TELEGRAM") {
		fmt.Fprintln(os.Stderr, "warning: LIFEOS_ALLOW_NO_TELEGRAM=true — bot polling/notifications disabled")
	}
	if envTruthy("LIFEOS_ALLOW_NO_API") {
		fmt.Fprintln(os.Stderr, "warning: LIFEOS_ALLOW_NO_API=true — REST / Mini App API not mounted")
	}
	if cfg.STTEnabled && strings.TrimSpace(cfg.STTAPIKey) == "" {
		fmt.Fprintln(os.Stderr, "warning: LIFEOS_STT_ENABLED=true but no STT/LLM API key — voice and video notes will be rejected")
	}
}

func rejectMockOllamaUnlessAllowed(cfg config.Config) error {
	if envTruthy("LIFEOS_ALLOW_MOCK_LLM") {
		fmt.Fprintln(os.Stderr, "warning: LIFEOS_ALLOW_MOCK_LLM=true — accepting mock / stub LLM (dev only)")
		return nil
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.OllamaURL), "/")
	if base == "" {
		base = "http://127.0.0.1:11434"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/tags", nil)
	if err != nil {
		return nil
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil // connectivity checked later; do not block boot on down Ollama
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var body struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil
	}
	for _, m := range body.Models {
		name := strings.ToLower(strings.TrimSpace(m.Name))
		if name == "lifeos_mock" || strings.HasPrefix(name, "lifeos_mock:") {
			return fmt.Errorf("Ollama at %s looks like scripts/mock_ollama.go (model lifeos_mock); set LIFEOS_ALLOW_MOCK_LLM=true for local stub testing, or run a real Ollama", base)
		}
	}
	return nil
}

func envTruthy(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
