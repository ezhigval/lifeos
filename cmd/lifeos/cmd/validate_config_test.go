package cmd

import (
	"strings"
	"testing"

	"github.com/valentinezhov/lifeos/internal/platform/config"
)

func TestValidateRuntimeConfigRequiresSecrets(t *testing.T) {
	t.Parallel()
	err := validateRuntimeConfig(config.Config{})
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "TELEGRAM_BOT_TOKEN") || !strings.Contains(msg, "LIFEOS_JWT_SECRET") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateRuntimeConfigOpenAINeedsKey(t *testing.T) {
	t.Parallel()
	err := validateRuntimeConfig(config.Config{
		TelegramBotToken: "t",
		JWTSecret:        "s",
		LLMEnabled:       true,
		LLMProvider:      "openai",
	})
	if err == nil || !strings.Contains(err.Error(), "LIFEOS_LLM_API_KEY") {
		t.Fatalf("got %v", err)
	}
}

func TestValidateRuntimeConfigOK(t *testing.T) {
	t.Parallel()
	if err := validateRuntimeConfig(config.Config{
		TelegramBotToken: "t",
		JWTSecret:        "s",
	}); err != nil {
		t.Fatal(err)
	}
}
