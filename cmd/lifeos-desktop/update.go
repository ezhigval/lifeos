package main

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxUpdateBytes = 80 << 20

type updateManifest struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
}

func versionNewer(latest, current string) bool {
	lp := versionParts(latest)
	cp := versionParts(current)
	n := len(lp)
	if len(cp) > n {
		n = len(cp)
	}
	for i := 0; i < n; i++ {
		var l, c int
		if i < len(lp) {
			l = lp[i]
		}
		if i < len(cp) {
			c = cp[i]
		}
		if l != c {
			return l > c
		}
	}
	return false
}

func versionParts(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if v == "" {
		return nil
	}
	raw := strings.Split(v, ".")
	out := make([]int, 0, len(raw))
	for _, part := range raw {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return nil
		}
		out = append(out, n)
	}
	return out
}

func sameOrigin(assetURL, manifestURL string) error {
	asset, err := url.Parse(assetURL)
	if err != nil {
		return err
	}
	manifest, err := url.Parse(manifestURL)
	if err != nil {
		return err
	}
	if asset.Host == "" || !strings.EqualFold(asset.Host, manifest.Host) || asset.Scheme != manifest.Scheme {
		return fmt.Errorf("update url")
	}
	if asset.Scheme != "https" && asset.Scheme != "http" {
		return fmt.Errorf("update url scheme")
	}
	return nil
}

func fetchManifest(ctx context.Context, client *http.Client, manifestURL string) (updateManifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return updateManifest{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		return updateManifest{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return updateManifest{}, fmt.Errorf("update manifest status %d", res.StatusCode)
	}
	var doc updateManifest
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&doc); err != nil {
		return updateManifest{}, err
	}
	doc.Version = strings.TrimSpace(doc.Version)
	doc.URL = strings.TrimSpace(doc.URL)
	doc.SHA256 = strings.ToLower(strings.TrimSpace(doc.SHA256))
	if doc.Version == "" || doc.URL == "" || len(doc.SHA256) != 64 {
		return updateManifest{}, fmt.Errorf("update manifest incomplete")
	}
	if err := sameOrigin(doc.URL, manifestURL); err != nil {
		return updateManifest{}, err
	}
	return doc, nil
}

func downloadVerified(ctx context.Context, client *http.Client, assetURL, wantSHA, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("update download status %d", res.StatusCode)
	}
	tmp, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer tmp.Close()
	hash := sha256.New()
	n, err := io.Copy(tmp, io.TeeReader(io.LimitReader(res.Body, maxUpdateBytes+1), hash))
	if err != nil {
		return err
	}
	if n > maxUpdateBytes {
		return fmt.Errorf("update too large")
	}
	sum := hex.EncodeToString(hash.Sum(nil))
	if sum != strings.ToLower(wantSHA) {
		return fmt.Errorf("update checksum mismatch")
	}
	return nil
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	root := filepath.Clean(dest)
	var written int64
	for _, f := range r.File {
		name := filepath.Clean(f.Name)
		if name == "." || strings.HasPrefix(name, "..") || filepath.IsAbs(f.Name) {
			return fmt.Errorf("update zip path")
		}
		target := filepath.Join(root, name)
		if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
			return fmt.Errorf("update zip path")
		}
		mode := f.Mode()
		if f.FileInfo().IsDir() || strings.HasSuffix(f.Name, "/") {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode&0o777)
		if err != nil {
			rc.Close()
			return err
		}
		n, err := io.Copy(out, io.LimitReader(rc, maxUpdateBytes+1))
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
		written += n
		if written > maxUpdateBytes {
			return fmt.Errorf("update too large")
		}
	}
	return nil
}

func appBundlePath() (string, bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", false
	}
	macOS := filepath.Dir(exe)
	if filepath.Base(macOS) != "MacOS" {
		return "", false
	}
	contents := filepath.Dir(macOS)
	if filepath.Base(contents) != "Contents" {
		return "", false
	}
	app := filepath.Dir(contents)
	if !strings.HasSuffix(app, ".app") {
		return "", false
	}
	return app, true
}

func watchUpdates(ctx context.Context, manifestURL, current, dataDir string, shutdown func()) {
	if strings.TrimSpace(manifestURL) == "" || strings.EqualFold(manifestURL, "off") {
		return
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	checkCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	manifest, err := fetchManifest(checkCtx, client, manifestURL)
	cancel()
	if err != nil {
		log.Printf("update check: %v", err)
		return
	}
	if !versionNewer(manifest.Version, current) {
		log.Printf("update check: %s is current", current)
		return
	}
	bundle, ok := appBundlePath()
	if !ok {
		log.Printf("update %s is available, this process is not inside LifeOS.app", manifest.Version)
		return
	}
	dir := filepath.Join(dataDir, "update")
	if err := os.RemoveAll(dir); err != nil {
		log.Printf("update dir: %v", err)
		return
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		log.Printf("update dir: %v", err)
		return
	}
	zipPath := filepath.Join(dir, "LifeOS-mac.zip")
	dlCtx, dlCancel := context.WithTimeout(ctx, 2*time.Minute)
	err = downloadVerified(dlCtx, client, manifest.URL, manifest.SHA256, zipPath)
	dlCancel()
	if err != nil {
		log.Printf("update download: %v", err)
		return
	}
	stage := filepath.Join(dir, "stage")
	if err := extractZip(zipPath, stage); err != nil {
		log.Printf("update extract: %v", err)
		return
	}
	staged := filepath.Join(stage, "LifeOS.app")
	if _, err := os.Stat(staged); errors.Is(err, os.ErrNotExist) {
		log.Printf("update zip has no LifeOS.app")
		return
	}
	log.Printf("update %s ready, restarting", manifest.Version)
	if err := stageAndRelaunch(staged, bundle); err != nil {
		log.Printf("update relaunch: %v", err)
		return
	}
	shutdown()
}
