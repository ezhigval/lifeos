package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/spheres/domain"
)

func TestSphereRename(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	s, err := domain.NewSphere(ids.NewUserID(), "Карьера", 1, now)
	if err != nil {
		t.Fatal(err)
	}

	later := now.Add(time.Hour)
	if err := s.Rename("  Карьера GO  ", 3, later); err != nil {
		t.Fatal(err)
	}
	if s.Name != "Карьера GO" || s.SortOrder != 3 {
		t.Fatalf("sphere = %+v", s)
	}
	if !s.UpdatedAt.Equal(later.UTC()) {
		t.Fatalf("UpdatedAt = %v", s.UpdatedAt)
	}

	if err := s.Rename("   ", 0, later); !errors.Is(err, domain.ErrEmptyName) {
		t.Fatalf("empty rename err = %v", err)
	}
}

func TestDefaultSphereNames(t *testing.T) {
	t.Parallel()
	if len(domain.DefaultSphereNames) == 0 {
		t.Fatal("expected default sphere names")
	}
	for _, name := range domain.DefaultSphereNames {
		if name == "" {
			t.Fatal("empty default name")
		}
	}
}
