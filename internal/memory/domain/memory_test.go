package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/memory/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestNewMemoryValidation(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	userID := ids.NewUserID()

	mem, err := domain.NewMemory(userID, domain.KindFact, "  home_city  ", "  Berlin  ", now)
	if err != nil {
		t.Fatal(err)
	}
	if mem.Key != "home_city" || mem.Value != "Berlin" {
		t.Fatalf("trimmed fields = %+v", mem)
	}
	if mem.Confidence != domain.DefaultConfidence || mem.Source != domain.DefaultSource {
		t.Fatalf("defaults = %+v", mem)
	}

	_, err = domain.NewMemory(ids.UserID{}, domain.KindFact, "k", "v", now)
	if err != domain.ErrInvalidUser {
		t.Fatalf("zero user: %v", err)
	}
	_, err = domain.NewMemory(userID, domain.Kind("x"), "k", "v", now)
	if err != domain.ErrInvalidKind {
		t.Fatalf("bad kind: %v", err)
	}
	_, err = domain.NewMemory(userID, domain.KindFact, " ", "v", now)
	if err != domain.ErrEmptyKey {
		t.Fatalf("empty key: %v", err)
	}
	_, err = domain.NewMemory(userID, domain.KindFact, "k", " ", now)
	if err != domain.ErrEmptyValue {
		t.Fatalf("empty value: %v", err)
	}
}
