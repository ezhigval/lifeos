package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/health/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestNewStepLog(t *testing.T) {
	t.Parallel()
	uid := ids.NewUserID()
	now := time.Now().UTC()
	log, err := domain.NewStepLog(uid, 8000, now, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.Steps != 8000 {
		t.Fatalf("steps = %d", log.Steps)
	}
}

func TestNewStepLogInvalid(t *testing.T) {
	t.Parallel()
	uid := ids.NewUserID()
	now := time.Now().UTC()
	_, err := domain.NewStepLog(uid, 0, now, now)
	if err != domain.ErrInvalidSteps {
		t.Fatalf("expected ErrInvalidSteps, got %v", err)
	}
	_, err = domain.NewStepLog(uid, 200001, now, now)
	if err != domain.ErrInvalidSteps {
		t.Fatalf("expected ErrInvalidSteps, got %v", err)
	}
}
