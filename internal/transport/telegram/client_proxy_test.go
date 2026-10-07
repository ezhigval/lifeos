package telegram

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientProxyUsesPlainHTTPOrigin(t *testing.T) {
	t.Setenv("LIFEOS_HTTP_PROXY", "http://127.0.0.1:8081")
	c := NewClient("42:TESTTOKEN")
	if c.origin != "http://api.telegram.org" {
		t.Fatalf("origin %s", c.origin)
	}
	if !strings.HasPrefix(c.base, "http://api.telegram.org/bot42:TESTTOKEN") {
		t.Fatalf("base %s", c.base)
	}
}

func TestNewClientDefaultIsHTTPS(t *testing.T) {
	t.Setenv("LIFEOS_HTTP_PROXY", "")
	c := NewClient("42:TESTTOKEN")
	if c.origin != "https://api.telegram.org" {
		t.Fatalf("origin %s", c.origin)
	}
}

func TestNewClientIgnoresProxyWithoutHost(t *testing.T) {
	t.Setenv("LIFEOS_HTTP_PROXY", "not a proxy")
	c := NewClient("42:TESTTOKEN")
	if c.origin != "https://api.telegram.org" {
		t.Fatalf("origin %s", c.origin)
	}
}

func TestProxyReceivesAbsoluteHTTPNotConnect(t *testing.T) {
	var gotMethod, gotURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true,"result":[]}`)
	}))
	t.Cleanup(srv.Close)

	t.Setenv("LIFEOS_HTTP_PROXY", srv.URL)
	c := NewClient("42:TESTTOKEN")
	if _, err := c.GetUpdates(context.Background(), 1, 0); err != nil {
		t.Fatal(err)
	}
	if gotMethod == http.MethodConnect {
		t.Fatalf("proxy saw CONNECT %s", gotURI)
	}
	if !strings.Contains(gotURI, "http://api.telegram.org/bot42:TESTTOKEN/getUpdates") {
		t.Fatalf("uri %s", gotURI)
	}
}

func TestGetUpdatesErrorRedactsToken(t *testing.T) {
	t.Setenv("LIFEOS_HTTP_PROXY", "http://127.0.0.1:1")
	const token = "42:TESTTOKEN"
	c := NewClient(token)
	_, err := c.GetUpdates(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), token) {
		t.Fatalf("token leaked: %s", err)
	}
}

func TestRedactEscapedToken(t *testing.T) {
	c := NewClient("42:TESTTOKEN")
	got := c.redact("http://api.telegram.org/bot42%3ATESTTOKEN/getMe")
	if strings.Contains(got, "TESTTOKEN") {
		t.Fatalf("escaped token leaked: %s", got)
	}
}
