package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/knowledge/domain"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestNewNoteRequiresBody(t *testing.T) {
	t.Parallel()
	_, err := domain.NewNote(ids.NewUserID(), "  ", nil, time.Now().UTC())
	if err != domain.ErrEmptyBody {
		t.Fatalf("err=%v", err)
	}
}
