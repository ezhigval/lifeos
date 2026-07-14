package otel_test

import (
	"context"
	"testing"

	"github.com/valentinezhov/lifeos/internal/platform/otel"
)

func TestInitDisabled(t *testing.T) {
	t.Parallel()

	shutdown, err := otel.Init(context.Background(), otel.Config{Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestInitEnabledRequiresEndpoint(t *testing.T) {
	t.Parallel()

	_, err := otel.Init(context.Background(), otel.Config{Enabled: true})
	if err == nil {
		t.Fatal("expected error without endpoint")
	}
}
