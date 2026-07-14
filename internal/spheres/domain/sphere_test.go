package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/spheres/domain"
)

func TestNewSphere(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	s, err := domain.NewSphere(ids.NewUserID(), "Карьера", 0, now)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if s.Name != "Карьера" {
		t.Fatalf("name=%q", s.Name)
	}
}

func TestNewSphereEmptyName(t *testing.T) {
	t.Parallel()
	_, err := domain.NewSphere(ids.NewUserID(), "  ", 0, time.Now().UTC())
	if err != domain.ErrEmptyName {
		t.Fatalf("err=%v", err)
	}
}
