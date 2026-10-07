package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestListenDesktopUsesStableLoopbackPort(t *testing.T) {
	ln, err := listenDesktop("")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	host, portText, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		t.Fatalf("host %s", host)
	}
	port, _ := strconv.Atoi(portText)
	if port < desktopPort || port >= desktopPort+desktopPortSpan {
		t.Fatalf("port %d", port)
	}

	next, err := listenDesktop("")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = next.Close() })
	_, nextPort, _ := net.SplitHostPort(next.Addr().String())
	if nextPort == portText {
		t.Fatalf("second listen reused %s", portText)
	}
}

func TestListenDesktopHonorsExplicitAddr(t *testing.T) {
	ln, err := listenDesktop("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	if ln.Addr().String() == "" {
		t.Fatal("empty addr")
	}
}

func TestSessionFileRoundTrip(t *testing.T) {
	path := t.TempDir() + "/session.json"
	store := openSessionStore(path)
	now := time.Unix(1_800_000_000, 0)
	token := fakeJWT(now.Add(2 * time.Hour).Unix())
	sess := storedSession{AccessToken: token, ExpiresAt: now.Add(2 * time.Hour).UnixMilli(), TelegramID: 42}
	if err := store.Save(sess); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", info.Mode().Perm())
	}
	got, ok := store.Load(now)
	if !ok || got.AccessToken != token || got.TelegramID != 42 {
		t.Fatalf("loaded %+v %v", got, ok)
	}
	if _, ok := store.Load(now.Add(3 * time.Hour)); ok {
		t.Fatal("expired session loaded")
	}
	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Load(now); ok {
		t.Fatal("cleared session loaded")
	}
}

func TestDesktopSessionSurvivesLoginAndRestart(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	token := fakeJWT(now.Add(24 * time.Hour).Unix())
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/telegram-login/verify" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":%q,"expires_in":86400,"token_type":"Bearer","telegram_id":7}`, token)
	}))
	t.Cleanup(origin.Close)

	dir := t.TempDir()
	cache, err := OpenCache(dir + "/cache.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cache.Close() })
	srv, err := newDesktopServer(testUI(), origin.URL, cache)
	if err != nil {
		t.Fatal(err)
	}
	srv.now = func() time.Time { return now }
	srv.sessions = openSessionStore(dir + "/session.json")
	desk := httptest.NewServer(srv.routes())
	t.Cleanup(desk.Close)

	res, err := http.Post(desk.URL+"/api/v1/auth/telegram-login/verify", "application/json", strings.NewReader(`{"username":"nick","code":"123456"}`))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("verify %d", res.StatusCode)
	}

	res, err = http.Get(desk.URL + "/desktop/session")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != 200 || res.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("session %d %s", res.StatusCode, body)
	}
	var got storedSession
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != token || got.TelegramID != 7 {
		t.Fatalf("stored %+v", got)
	}
	if strings.Contains(string(body), "123456") {
		t.Fatal("login code stored")
	}

	// A new process with the same file still has the login.
	again := openSessionStore(dir + "/session.json")
	loaded, ok := again.Load(now)
	if !ok || loaded.AccessToken != token {
		t.Fatalf("restart %+v %v", loaded, ok)
	}

	req, _ := http.NewRequest(http.MethodDelete, desk.URL+"/desktop/session", nil)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete %d", res.StatusCode)
	}
	res, err = http.Get(desk.URL + "/desktop/session")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("after delete %d", res.StatusCode)
	}
}

func TestFailedLoginDoesNotStoreSession(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"bad"}`, http.StatusUnauthorized)
	}))
	t.Cleanup(origin.Close)
	dir := t.TempDir()
	cache, err := OpenCache(dir + "/cache.db")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := newDesktopServer(testUI(), origin.URL, cache)
	if err != nil {
		t.Fatal(err)
	}
	srv.sessions = openSessionStore(dir + "/session.json")
	desk := httptest.NewServer(srv.routes())
	t.Cleanup(desk.Close)

	res, err := http.Post(desk.URL+"/api/v1/auth/telegram-login/verify", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if _, err := os.Stat(dir + "/session.json"); !os.IsNotExist(err) {
		t.Fatalf("session file err %v", err)
	}
}

func fakeJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, exp)))
	return header + "." + payload + ".sig"
}
