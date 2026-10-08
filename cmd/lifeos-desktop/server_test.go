package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"
)

func testUI() fsTest {
	return fstest.MapFS{
		"index.html":    {Data: []byte("<!doctype html><title>LifeOS</title>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
	}
}

type fsTest = fstest.MapFS

func TestProxyCachesGetAndFallsBackWhenOriginIsDown(t *testing.T) {
	var hits atomic.Int32
	var fail atomic.Bool
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if fail.Load() {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		if r.URL.Path != "/api/v1/tasks/today" {
			t.Errorf("path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Errorf("auth %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"tasks":[]}`)
	}))
	t.Cleanup(origin.Close)

	cache, err := OpenCache(t.TempDir() + "/cache.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cache.Close() })

	srv, err := newDesktopServer(testUI(), origin.URL, cache)
	if err != nil {
		t.Fatal(err)
	}
	desk := httptest.NewServer(srv.routes())
	t.Cleanup(desk.Close)

	req, _ := http.NewRequest(http.MethodGet, desk.URL+"/api/v1/tasks/today", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != 200 || string(body) != `{"tasks":[]}` {
		t.Fatalf("first %d %s", res.StatusCode, body)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits %d", hits.Load())
	}

	req, _ = http.NewRequest(http.MethodGet, desk.URL+"/api/v1/tasks/today", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.Header.Get("X-LifeOS-Cache") != "hit" || string(body) != `{"tasks":[]}` {
		t.Fatalf("cached %s %s", res.Header.Get("X-LifeOS-Cache"), body)
	}
	if hits.Load() != 1 {
		t.Fatalf("second request reached origin, hits %d", hits.Load())
	}

	var stored string
	if err := cache.db.QueryRow(`SELECT cache_key FROM http_cache`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored, "secret-token") {
		t.Fatal("token stored in cache key")
	}

	fail.Store(true)
	srv.now = func() time.Time { return time.Now().Add(10 * time.Minute) }
	req, _ = http.NewRequest(http.MethodGet, desk.URL+"/api/v1/tasks/today", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != 200 || string(body) != `{"tasks":[]}` || res.Header.Get("X-LifeOS-Cache") != "hit" {
		t.Fatalf("stale %d %s cache %s", res.StatusCode, body, res.Header.Get("X-LifeOS-Cache"))
	}
}

func TestPostIsForwardedAndNotCached(t *testing.T) {
	var gotBody string
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(origin.Close)

	cache, err := OpenCache(t.TempDir() + "/cache.db")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := newDesktopServer(testUI(), origin.URL, cache)
	if err != nil {
		t.Fatal(err)
	}
	desk := httptest.NewServer(srv.routes())
	t.Cleanup(desk.Close)

	res, err := http.Post(desk.URL+"/api/v1/auth/telegram-login/request", "application/json", strings.NewReader(`{"username":"nick"}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusCreated || gotBody != `{"username":"nick"}` {
		t.Fatalf("status %d body %s", res.StatusCode, gotBody)
	}
	var n int
	if err := cache.db.QueryRow(`SELECT COUNT(*) FROM http_cache`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("post cached rows %d", n)
	}
}

func TestAssistantChatProxyIgnoresShortAPITimeout(t *testing.T) {
	if chatProxyTimeout <= apiProxyTimeout {
		t.Fatalf("chat timeout %s must exceed api timeout %s", chatProxyTimeout, apiProxyTimeout)
	}
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != assistantChatPath {
			t.Errorf("path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Errorf("auth %q", got)
		}
		time.Sleep(30 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"reply":"ок","waiting":false,"history":[],"tools_run":[]}`)
	}))
	t.Cleanup(origin.Close)

	cache, err := OpenCache(t.TempDir() + "/cache.db")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := newDesktopServer(testUI(), origin.URL, cache)
	if err != nil {
		t.Fatal(err)
	}
	srv.client = &http.Client{Timeout: time.Millisecond}
	desk := httptest.NewServer(srv.routes())
	t.Cleanup(desk.Close)

	req, _ := http.NewRequest(http.MethodPost, desk.URL+assistantChatPath, strings.NewReader(`{"text":"привет"}`))
	req.Header.Set("Authorization", "Bearer secret-token")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `"ок"`) {
		t.Fatalf("status %d body %s", res.StatusCode, body)
	}
	var n int
	if err := cache.db.QueryRow(`SELECT COUNT(*) FROM http_cache`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("chat post cached rows %d", n)
	}
}

func TestUIFallsBackToIndexForRoutes(t *testing.T) {
	cache, err := OpenCache(t.TempDir() + "/cache.db")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := newDesktopServer(testUI(), "https://example.test", cache)
	if err != nil {
		t.Fatal(err)
	}
	desk := httptest.NewServer(srv.routes())
	t.Cleanup(desk.Close)

	res, err := http.Get(desk.URL + "/more/notes")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if !strings.Contains(string(body), "<title>LifeOS</title>") {
		t.Fatalf("body %s", body)
	}

	res, err = http.Get(desk.URL + "/assets/app.js")
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	if string(body) != "console.log(1)" {
		t.Fatalf("js %s", body)
	}

	res, err = http.Get(desk.URL + "/assets/missing.js")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("missing asset %d", res.StatusCode)
	}
}
