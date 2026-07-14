package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/health/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestNewWeightLogRejectsOutOfRange(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	_, err := domain.NewWeightLog(ids.NewUserID(), 10, now, now)
	if err != domain.ErrInvalidWeight {
		t.Fatalf("err=%v", err)
	}
}
