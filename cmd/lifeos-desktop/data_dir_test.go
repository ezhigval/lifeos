package main

import (
	"path/filepath"
	"testing"
)

func TestDesktopDataDir(t *testing.T) {
	t.Parallel()
	home := filepath.Join("home", "a")
	if got := desktopDataDir("linux", home, "", "/override"); got != "/override" {
		t.Fatalf("override %s", got)
	}
	if got := desktopDataDir("linux", "", "", ""); got != ".lifeos" {
		t.Fatalf("empty home %s", got)
	}
	if got := desktopDataDir("darwin", home, "", ""); got != filepath.Join(home, "Library", "Application Support", "LifeOS") {
		t.Fatalf("darwin %s", got)
	}
	if got := desktopDataDir("linux", home, "", ""); got != filepath.Join(home, ".local", "share", "lifeos") {
		t.Fatalf("linux %s", got)
	}
	local := filepath.Join(home, "AppData", "Local")
	if got := desktopDataDir("windows", home, local, ""); got != filepath.Join(local, "LifeOS") {
		t.Fatalf("windows local %s", got)
	}
	if got := desktopDataDir("windows", home, "", ""); got != filepath.Join(home, "AppData", "Local", "LifeOS") {
		t.Fatalf("windows fallback %s", got)
	}
}
