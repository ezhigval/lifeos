package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/valentinezhov/lifeos/internal/platform/ids"
	"github.com/valentinezhov/lifeos/internal/projects/domain"
)

func TestNewProjectValidation(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	now := time.Now().UTC()

	if _, err := domain.NewProject(userID, "", []ids.SphereID{ids.NewSphereID()}, now); !errors.Is(err, domain.ErrEmptyName) {
		t.Fatalf("empty name err = %v", err)
	}
	if _, err := domain.NewProject(userID, "proj", nil, now); !errors.Is(err, domain.ErrNoSphere) {
		t.Fatalf("no sphere err = %v", err)
	}

	p, err := domain.NewProject(userID, "свадьба", []ids.SphereID{ids.NewSphereID()}, now)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != domain.StatusActive || !p.CurrentValue.IsZero() {
		t.Fatalf("project = %+v", p)
	}
}

func TestStatusValid(t *testing.T) {
	t.Parallel()
	for _, s := range []domain.Status{domain.StatusActive, domain.StatusArchived, domain.StatusCompleted} {
		if !s.Valid() {
			t.Fatalf("%s should be valid", s)
		}
	}
	if domain.Status("draft").Valid() {
		t.Fatal("draft should be invalid")
	}
}

func TestProjectProgress(t *testing.T) {
	t.Parallel()
	userID := ids.NewUserID()
	p, err := domain.NewProject(userID, "goal", []ids.SphereID{ids.NewSphereID()}, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Progress()
	if !errors.Is(err, domain.ErrNoTarget) {
		t.Fatalf("no target err = %v", err)
	}

	target := decimal.NewFromInt(100)
	p.TargetValue = &target
	p.CurrentValue = decimal.NewFromInt(40)
	unit := "ч"
	p.Unit = &unit

	prog, err := p.Progress()
	if err != nil {
		t.Fatal(err)
	}
	if !prog.HasTarget || prog.Unit != "ч" {
		t.Fatalf("prog = %+v", prog)
	}
	if !prog.Remaining.Equal(decimal.NewFromInt(60)) {
		t.Fatalf("remaining = %s", prog.Remaining)
	}
	if !prog.Percent.Equal(decimal.NewFromInt(40)) {
		t.Fatalf("percent = %s", prog.Percent)
	}

	p.CurrentValue = decimal.NewFromInt(150)
	prog, err = p.Progress()
	if err != nil {
		t.Fatal(err)
	}
	if !prog.Remaining.IsZero() {
		t.Fatalf("over-target remaining = %s", prog.Remaining)
	}

	zero := decimal.Zero
	p.TargetValue = &zero
	p.CurrentValue = decimal.NewFromInt(10)
	prog, err = p.Progress()
	if err != nil {
		t.Fatal(err)
	}
	if !prog.Percent.IsZero() {
		t.Fatalf("zero target percent = %s", prog.Percent)
	}
}
