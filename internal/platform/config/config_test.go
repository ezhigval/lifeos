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
}
