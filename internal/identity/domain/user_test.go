package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/identity/domain"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	user, err := domain.NewUser(12345, "Valentin", "Europe/Moscow", now)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	if user.TelegramID != 12345 {
		t.Fatalf("TelegramID = %d, want 12345", user.TelegramID)
	}
	if user.DisplayName != "Valentin" {
		t.Fatalf("DisplayName = %q", user.DisplayName)
	}
	if user.ID.IsZero() {
		t.Fatal("expected generated user id")
	}
}

func TestNewUserValidation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	_, err := domain.NewUser(0, "Valentin", "Europe/Moscow", now)
	if err == nil {
		t.Fatal("expected error for invalid telegram id")
	}

	_, err = domain.NewUser(1, "", "Europe/Moscow", now)
	if err == nil {
		t.Fatal("expected error for empty display name")
	}
}
