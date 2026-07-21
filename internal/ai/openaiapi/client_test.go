package openaiapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientChatSuccess(t *testing.T) {
	t.Parallel()
	var gotAuth, gotPath string
	var gotBody chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":" hello "}}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL+"/", "test-key", "test-model")
	got, err := c.Chat(context.Background(), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("auth %q", gotAuth)
	}
	if gotPath != "/chat/completions" {
		t.Fatalf("path %q", gotPath)
	}
	if gotBody.Model != "test-model" || gotBody.Temperature != 0 {
		t.Fatalf("body %+v", gotBody)
	}
	if gotBody.ResponseFormat != nil {
		t.Fatalf("Chat should not set response_format")
	}
}

func TestClientChatErrorStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"busy"}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "k", "m")
	_, err := c.Chat(context.Background(), "s", "u")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Fatalf("err %v", err)
	}
}

func TestClientChatJSONRetriesWithoutResponseFormat(t *testing.T) {
	t.Parallel()
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		raw, _ := io.ReadAll(r.Body)
		var body chatRequest
		_ = json.Unmarshal(raw, &body)
		if body.ResponseFormat != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"response_format not supported"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"intent\":\"unknown\"}"}}]}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL, "k", "m")
	got, err := c.ChatJSON(context.Background(), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "unknown") {
		t.Fatalf("got %q", got)
	}
	if calls != 2 {
		t.Fatalf("calls=%d want 2", calls)
	}
}

func TestNewClientDefaults(t *testing.T) {
	t.Parallel()
	c := NewClient("", "", "")
	if c.baseURL != defaultBaseURL {
		t.Fatalf("baseURL %q", c.baseURL)
	}
	if c.model != defaultModel {
		t.Fatalf("model %q", c.model)
	}
}
