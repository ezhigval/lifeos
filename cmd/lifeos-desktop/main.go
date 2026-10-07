package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

//go:embed all:ui
var uiEmbed embed.FS

// desktopVersion is overridden by -X main.desktopVersion in the Mac build.
var desktopVersion = "0.2.0"

func main() {
	log.SetFlags(0)
	ui, err := fs.Sub(uiEmbed, "ui")
	if err != nil {
		log.Fatal(err)
	}
	origin := os.Getenv("LIFEOS_API_ORIGIN")
	if origin == "" {
		origin = "https://local-ai-assist.ru"
	}
	cache, err := OpenCache(filepath.Join(dataDir(), "cache.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	desktop, err := newDesktopServer(ui, origin, cache)
	if err != nil {
		log.Fatal(err)
	}

	addr := os.Getenv("LIFEOS_DESKTOP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	host, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		log.Fatal("desktop server refuses a non-loopback address")
	}

	httpSrv := &http.Server{
		Handler:           desktop.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shut, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shut)
	}()

	// The Mac window reads this single line from stdout.
	fmt.Printf("PORT=%s\n", port)
	log.Printf("lifeos desktop %s http://127.0.0.1:%s api %s", desktopVersion, port, origin)
	go watchUpdates(ctx, updateURL(), desktopVersion, dataDir(), func() {
		shut, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shut)
	})
	if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// githubReleaseManifest is the desktop update feed. The zip and this JSON are
// GitHub Release assets from .github/workflows/desktop-release.yml. The VM
// does not serve them, and that workflow has no R2 bucket.
const githubReleaseManifest = "https://github.com/ezhigval/lifeos/releases/latest/download/latest.json"

func updateURL() string {
	if v := os.Getenv("LIFEOS_UPDATE_URL"); v != "" {
		return v
	}
	return githubReleaseManifest
}

func dataDir() string {
	if v := os.Getenv("LIFEOS_DESKTOP_DATA"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".lifeos"
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "LifeOS")
	}
	return filepath.Join(home, ".local", "share", "lifeos")
}
