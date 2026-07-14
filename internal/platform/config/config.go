package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
	LogLevel    string
	LogFormat   string

	TelegramBotToken      string
	TelegramMode          string
	TelegramWebhookURL    string
	TelegramWebhookSecret string

	SeedTelegramID  int64
	SeedDisplayName string
	SeedTimezone    string
	MigrationsDir   string

	OtelEnabled  bool
	OtelEndpoint string

	LLMEnabled  bool
	OllamaURL   string
	OllamaModel string

	JWTSecret   string
	APIKey      string
	JWTTTLHours int

	// MiniAppURL is the public HTTPS URL of the Telegram Mini App (…/app/).
	MiniAppURL string
	// StaticDir serves built Mini App files at /app/ when non-empty.
	StaticDir string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:           envOr("LIFEOS_DATABASE_URL", "postgres://lifeos:lifeos@localhost:5433/lifeos?sslmode=disable"),
		HTTPAddr:              envOr("LIFEOS_HTTP_ADDR", ":8080"),
		LogLevel:              envOr("LIFEOS_LOG_LEVEL", "info"),
		LogFormat:             envOr("LIFEOS_LOG_FORMAT", "text"),
		TelegramBotToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramMode:          envOr("LIFEOS_TELEGRAM_MODE", "polling"),
		TelegramWebhookURL:    os.Getenv("LIFEOS_TELEGRAM_WEBHOOK_URL"),
		TelegramWebhookSecret: os.Getenv("LIFEOS_TELEGRAM_WEBHOOK_SECRET"),
		SeedDisplayName:       envOr("LIFEOS_SEED_DISPLAY_NAME", "Valentin"),
		SeedTimezone:          envOr("LIFEOS_SEED_TIMEZONE", "Europe/Moscow"),
		MigrationsDir:         envOr("LIFEOS_MIGRATIONS_DIR", "migrations"),
		OtelEndpoint:          envOr("LIFEOS_OTEL_ENDPOINT", "localhost:4318"),
	}

	otelEnabled, err := parseBoolDefault(os.Getenv("LIFEOS_OTEL_ENABLED"), false)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_OTEL_ENABLED: %w", err)
	}
	cfg.OtelEnabled = otelEnabled

	llmEnabled, err := parseBoolDefault(os.Getenv("LIFEOS_LLM_ENABLED"), false)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_LLM_ENABLED: %w", err)
	}
	cfg.LLMEnabled = llmEnabled
	cfg.OllamaURL = envOr("LIFEOS_OLLAMA_URL", "http://localhost:11434")
	cfg.OllamaModel = envOr("LIFEOS_OLLAMA_MODEL", "llama3.2")
	cfg.JWTSecret = os.Getenv("LIFEOS_JWT_SECRET")
	cfg.APIKey = os.Getenv("LIFEOS_API_KEY")
	cfg.MiniAppURL = strings.TrimSpace(os.Getenv("LIFEOS_MINIAPP_URL"))
	cfg.StaticDir = envOr("LIFEOS_STATIC_DIR", "web/miniapp/dist")
	ttlHours, err := parseInt64Default(os.Getenv("LIFEOS_JWT_TTL_HOURS"), 24)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_JWT_TTL_HOURS: %w", err)
	}
	cfg.JWTTTLHours = int(ttlHours)

	seedID, err := parseInt64Default(os.Getenv("LIFEOS_SEED_TELEGRAM_ID"), 0)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_SEED_TELEGRAM_ID: %w", err)
	}
	cfg.SeedTelegramID = seedID

	if cfg.LogFormat != "text" && cfg.LogFormat != "json" {
		return Config{}, fmt.Errorf("LIFEOS_LOG_FORMAT: want text or json, got %q", cfg.LogFormat)
	}
	if cfg.TelegramMode == "webhook" {
		if cfg.TelegramBotToken == "" {
			return Config{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required for webhook mode")
		}
		if cfg.TelegramWebhookURL == "" {
			return Config{}, fmt.Errorf("LIFEOS_TELEGRAM_WEBHOOK_URL is required for webhook mode")
		}
		if cfg.TelegramWebhookSecret == "" {
			return Config{}, fmt.Errorf("LIFEOS_TELEGRAM_WEBHOOK_SECRET is required for webhook mode")
		}
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt64List(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	out := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid id %q: %w", part, err)
		}
		out = append(out, n)
	}
	return out, nil
}

func parseInt64Default(raw string, fallback int64) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func parseBoolDefault(raw string, fallback bool) (bool, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return fallback, nil
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %q", raw)
	}
}
