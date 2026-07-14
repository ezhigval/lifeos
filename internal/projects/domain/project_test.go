package domain_test

import (
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/projects/domain"
)

func TestProjectArchive(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	p, err := domain.NewProject(userID, "свадьба", []ids.SphereID{ids.NewSphereID()}, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Archive(); err != nil {
		t.Fatal(err)
	}
	if p.Status != domain.StatusArchived {
		t.Fatalf("status = %s, want archived", p.Status)
	}
	if err := p.Archive(); err != domain.ErrNotActive {
		t.Fatalf("second archive err = %v", err)
	}
}
