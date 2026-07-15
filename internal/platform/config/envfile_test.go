package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFilesDoesNotOverrideExisting(t *testing.T) {
	dir := t.TempDir()
	prev, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prev) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	content := "LIFEOS_LOG_LEVEL=debug\nLIFEOS_HTTP_ADDR=:9999\n"
	if err := os.WriteFile(filepath.Join(dir, "settings.env"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("LIFEOS_LOG_LEVEL", "warn")
	_ = os.Unsetenv("LIFEOS_HTTP_ADDR")

	loadEnvFiles()

	if got := os.Getenv("LIFEOS_LOG_LEVEL"); got != "warn" {
		t.Fatalf("existing env should win, got %q", got)
	}
	if got := os.Getenv("LIFEOS_HTTP_ADDR"); got != ":9999" {
		t.Fatalf("settings.env should fill missing, got %q", got)
	}
}
