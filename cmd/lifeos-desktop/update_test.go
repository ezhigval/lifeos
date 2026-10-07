package main

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateURLIsGitHubRelease(t *testing.T) {
	t.Setenv("LIFEOS_UPDATE_URL", "")
	got := updateURL()
	if got != githubReleaseManifest {
		t.Fatalf("default update url %s", got)
	}
	if strings.Contains(got, "/app/desktop/") {
		t.Fatal("updater must not use the VM mini app path")
	}
	t.Setenv("LIFEOS_UPDATE_URL", "https://example.com/latest.json")
	if updateURL() != "https://example.com/latest.json" {
		t.Fatal("LIFEOS_UPDATE_URL override")
	}
}

func TestVersionNewer(t *testing.T) {
	if !versionNewer("0.2.0", "0.1.0") {
		t.Fatal("0.2.0 should replace 0.1.0")
	}
	if versionNewer("0.2.0", "0.2.0") || versionNewer("0.1.9", "0.2.0") {
		t.Fatal("older or same version must stay")
	}
	if !versionNewer("v0.2.1", "0.2.0") {
		t.Fatal("v prefix")
	}
	if versionNewer("nope", "0.2.0") {
		t.Fatal("garbage version")
	}
}

func TestUpdateRejectsOtherHost(t *testing.T) {
	manifest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"version":"9.0.0","url":"https://evil.example/app.zip","sha256":"` + hex64('a') + `"}`))
	}))
	t.Cleanup(manifest.Close)
	_, err := fetchManifest(context.Background(), manifest.Client(), manifest.URL+"/desktop/latest.json")
	if err == nil {
		t.Fatal("expected host rejection")
	}
}

func TestUpdateDownloadsMatchingZip(t *testing.T) {
	body := []byte("zip-bytes")
	sum := sha256.Sum256(body)
	sha := hex.EncodeToString(sum[:])
	var origin *httptest.Server
	origin = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/desktop/latest.json":
			_, _ = w.Write([]byte(`{"version":"0.2.1","url":"` + origin.URL + `/desktop/LifeOS-mac.zip","sha256":"` + sha + `"}`))
		default:
			_, _ = w.Write(body)
		}
	}))
	t.Cleanup(origin.Close)

	doc, err := fetchManifest(context.Background(), origin.Client(), origin.URL+"/desktop/latest.json")
	if err != nil {
		t.Fatal(err)
	}
	if !versionNewer(doc.Version, "0.2.0") {
		t.Fatal("expected newer")
	}
	dest := filepath.Join(t.TempDir(), "app.zip")
	if err := downloadVerified(context.Background(), origin.Client(), doc.URL, doc.SHA256, dest); err != nil {
		t.Fatal(err)
	}
	if err := downloadVerified(context.Background(), origin.Client(), doc.URL, hex64('b'), dest); err == nil {
		t.Fatal("expected checksum mismatch")
	}
}

func TestExtractZipRejectsSlip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "bad.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("../escape.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("nope")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := extractZip(zipPath, filepath.Join(dir, "out")); err == nil {
		t.Fatal("expected zip slip rejection")
	}
}

func hex64(b byte) string {
	return hex.EncodeToString([]byte{
		b, b, b, b, b, b, b, b,
		b, b, b, b, b, b, b, b,
		b, b, b, b, b, b, b, b,
		b, b, b, b, b, b, b, b,
	})
}
