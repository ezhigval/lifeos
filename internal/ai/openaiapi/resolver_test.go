package openaiapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestResolverWithFakeRoundTripper(t *testing.T) {
	t.Parallel()
	c := NewClient("https://api.example.com/v1", "key", "model")
	c.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer key" {
			t.Fatalf("auth %q", got)
		}
		body := `{"choices":[{"message":{"content":"{\"intent\":\"reminder.create\",\"message\":\"выпить воду\",\"time_text\":\"через час\",\"confidence\":0.95}"}}]}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Request:    r,
		}, nil
	})

	r := NewResolver(c)
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "напомни через час выпить воду"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentReminderCreate {
		t.Fatalf("type %s", got.Type)
	}
	if got.Message != "выпить воду" || got.TimeText != "через час" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolverDegradesOnHTTPError(t *testing.T) {
	t.Parallel()
	c := NewClient("https://api.example.com/v1", "key", "model")
	c.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(`{"error":"busy"}`)),
			Request:    r,
		}, nil
	})

	res := NewResolver(c)
	got, err := res.Resolve(context.Background(), ai.ResolveInput{Text: "что-то"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s", got.Type)
	}
}
