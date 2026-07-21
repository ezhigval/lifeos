package config_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/platform/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("LIFEOS_DATABASE_URL", "")
	t.Setenv("LIFEOS_HTTP_ADDR", "")
	t.Setenv("LIFEOS_LOG_LEVEL", "")
	t.Setenv("LIFEOS_LOG_FORMAT", "text")
	t.Setenv("LIFEOS_LLM_ENABLED", "")
	t.Setenv("LIFEOS_LLM_PROVIDER", "")
	t.Setenv("LIFEOS_LLM_API_KEY", "")
	t.Setenv("LIFEOS_OPENAI_API_KEY", "")
	t.Setenv("LIFEOS_LLM_BASE_URL", "")
	t.Setenv("LIFEOS_LLM_MODEL", "")
	t.Setenv("LIFEOS_OLLAMA_URL", "")
	t.Setenv("LIFEOS_OLLAMA_MODEL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.LogFormat != "text" {
		t.Fatalf("LogFormat = %q", cfg.LogFormat)
	}
	if cfg.LLMEnabled {
		t.Fatalf("LLMEnabled = true, want false")
	}
	if cfg.LLMProvider != "openai" {
		t.Fatalf("LLMProvider = %q, want openai", cfg.LLMProvider)
	}
	if cfg.LLMBaseURL != "https://api.groq.com/openai/v1" {
		t.Fatalf("LLMBaseURL = %q", cfg.LLMBaseURL)
	}
	if cfg.LLMModel != "llama-3.3-70b-versatile" {
		t.Fatalf("LLMModel = %q", cfg.LLMModel)
	}
	if cfg.OllamaURL != "http://localhost:11434" {
		t.Fatalf("OllamaURL = %q", cfg.OllamaURL)
	}
	if cfg.OllamaModel != "llama3.2" {
		t.Fatalf("OllamaModel = %q", cfg.OllamaModel)
	}
}

func TestLoadLLMAPIKeyAlias(t *testing.T) {
	t.Setenv("LIFEOS_LOG_FORMAT", "text")
	t.Setenv("LIFEOS_LLM_API_KEY", "")
	t.Setenv("LIFEOS_OPENAI_API_KEY", "sk-alias")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLMAPIKey != "sk-alias" {
		t.Fatalf("LLMAPIKey = %q, want sk-alias", cfg.LLMAPIKey)
	}
}

func TestLoadLLMAPIKeyPrefersPrimary(t *testing.T) {
	t.Setenv("LIFEOS_LOG_FORMAT", "text")
	t.Setenv("LIFEOS_LLM_API_KEY", "sk-primary")
	t.Setenv("LIFEOS_OPENAI_API_KEY", "sk-alias")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLMAPIKey != "sk-primary" {
		t.Fatalf("LLMAPIKey = %q, want sk-primary", cfg.LLMAPIKey)
	}
}

func TestLoadLLMProviderOllama(t *testing.T) {
	t.Setenv("LIFEOS_LOG_FORMAT", "text")
	t.Setenv("LIFEOS_LLM_PROVIDER", "OLLAMA")
	t.Setenv("LIFEOS_OLLAMA_URL", "http://ollama:11434")
	t.Setenv("LIFEOS_OLLAMA_MODEL", "llama3.1")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.LLMProvider != "ollama" {
		t.Fatalf("LLMProvider = %q, want ollama", cfg.LLMProvider)
	}
	if cfg.OllamaURL != "http://ollama:11434" {
		t.Fatalf("OllamaURL = %q", cfg.OllamaURL)
	}
	if cfg.OllamaModel != "llama3.1" {
		t.Fatalf("OllamaModel = %q", cfg.OllamaModel)
	}
}
