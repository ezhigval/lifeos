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

	LLMEnabled      bool
	LLMAgentEnabled bool   // multi-turn conversational agent (tools + dialogue)
	LLMProvider     string // "ollama" | "openai" (default)
	LLMAPIKey       string
	LLMBaseURL      string // OpenAI-compatible base, e.g. https://api.groq.com/openai/v1
	LLMModel        string // model for openai provider; if empty use provider default
	OllamaURL       string
	OllamaModel     string
	LearningSalt    string // HMAC salt for anonymous learning subjects

	// Speech-to-text for Telegram voice / video notes (OpenAI-compatible Whisper).
	STTEnabled bool
	STTAPIKey  string
	STTBaseURL string
	STTModel   string

	// Vision for Telegram photos / image documents without caption.
	VisionEnabled bool
	VisionAPIKey  string
	VisionBaseURL string
	VisionModel   string

	JWTSecret   string
	APIKey      string
	JWTTTLHours int
	// WebAppAuthTTLHours is how long Telegram initData remains acceptable (HMAC auth_date).
	// Independent of JWT session TTL — keep short (default 24h) even if Mini App JWT is longer.
	WebAppAuthTTLHours int

	// MiniAppURL is the public HTTPS URL of the Telegram Mini App (…/app/).
	MiniAppURL string
	// StaticDir serves built Mini App files at /app/ when non-empty.
	StaticDir string
}

func Load() (Config, error) {
	loadEnvFiles()

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
	llmAgent, err := parseBoolDefault(os.Getenv("LIFEOS_LLM_AGENT_ENABLED"), true)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_LLM_AGENT_ENABLED: %w", err)
	}
	cfg.LLMAgentEnabled = llmAgent
	cfg.LLMProvider = strings.ToLower(strings.TrimSpace(envOr("LIFEOS_LLM_PROVIDER", "openai")))
	cfg.LLMAPIKey = firstNonEmpty(os.Getenv("LIFEOS_LLM_API_KEY"), os.Getenv("LIFEOS_OPENAI_API_KEY"))
	cfg.LLMBaseURL = envOr("LIFEOS_LLM_BASE_URL", "https://api.groq.com/openai/v1")
	cfg.LLMModel = envOr("LIFEOS_LLM_MODEL", "llama-3.3-70b-versatile")
	cfg.OllamaURL = envOr("LIFEOS_OLLAMA_URL", "http://localhost:11434")
	cfg.OllamaModel = envOr("LIFEOS_OLLAMA_MODEL", "llama3.2")
	cfg.LearningSalt = envOr("LIFEOS_LEARNING_SALT", "lifeos-dev-learning-salt-change-me")

	sttEnabled, err := parseBoolDefault(os.Getenv("LIFEOS_STT_ENABLED"), false)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_STT_ENABLED: %w", err)
	}
	cfg.STTEnabled = sttEnabled
	cfg.STTAPIKey = firstNonEmpty(os.Getenv("LIFEOS_STT_API_KEY"), cfg.LLMAPIKey)
	cfg.STTBaseURL = envOr("LIFEOS_STT_BASE_URL", cfg.LLMBaseURL)
	cfg.STTModel = envOr("LIFEOS_STT_MODEL", "whisper-large-v3-turbo")

	visionEnabled, err := parseBoolDefault(os.Getenv("LIFEOS_VISION_ENABLED"), false)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_VISION_ENABLED: %w", err)
	}
	cfg.VisionEnabled = visionEnabled
	cfg.VisionAPIKey = firstNonEmpty(os.Getenv("LIFEOS_VISION_API_KEY"), cfg.LLMAPIKey)
	cfg.VisionBaseURL = envOr("LIFEOS_VISION_BASE_URL", cfg.LLMBaseURL)
	cfg.VisionModel = envOr("LIFEOS_VISION_MODEL", "meta-llama/llama-4-scout-17b-16e-instruct")

	cfg.JWTSecret = os.Getenv("LIFEOS_JWT_SECRET")
	cfg.APIKey = os.Getenv("LIFEOS_API_KEY")
	cfg.MiniAppURL = strings.TrimSpace(os.Getenv("LIFEOS_MINIAPP_URL"))
	cfg.StaticDir = envOr("LIFEOS_STATIC_DIR", "web/miniapp/dist")
	ttlHours, err := parseInt64Default(os.Getenv("LIFEOS_JWT_TTL_HOURS"), 24)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_JWT_TTL_HOURS: %w", err)
	}
	cfg.JWTTTLHours = int(ttlHours)
	webAppTTL, err := parseInt64Default(os.Getenv("LIFEOS_WEBAPP_AUTH_TTL_HOURS"), 24)
	if err != nil {
		return Config{}, fmt.Errorf("LIFEOS_WEBAPP_AUTH_TTL_HOURS: %w", err)
	}
	cfg.WebAppAuthTTLHours = int(webAppTTL)

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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
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
