package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/health/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestNewSleepLog(t *testing.T) {
	t.Parallel()
	uid := ids.NewUserID()
	now := time.Now().UTC()
	log, err := domain.NewSleepLog(uid, 450, now, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.DurationMinutes != 450 {
		t.Fatalf("duration = %d", log.DurationMinutes)
	}
}

func TestNewSleepLogInvalid(t *testing.T) {
	t.Parallel()
	uid := ids.NewUserID()
	now := time.Now().UTC()
	_, err := domain.NewSleepLog(uid, 0, now, now)
	if err != domain.ErrInvalidSleep {
		t.Fatalf("expected ErrInvalidSleep, got %v", err)
	}
	_, err = domain.NewSleepLog(uid, 1441, now, now)
	if err != domain.ErrInvalidSleep {
		t.Fatalf("expected ErrInvalidSleep, got %v", err)
	}
}
