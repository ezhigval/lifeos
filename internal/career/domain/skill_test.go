package domain_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/career/domain"
)

func TestParseSkillLine(t *testing.T) {
	t.Parallel()
	name, level := domain.ParseSkillLine("Go — senior")
	if name != "Go" || level != "senior" {
		t.Fatalf("got %q %q", name, level)
	}
	name, level = domain.ParseSkillLine("PostgreSQL middle")
	if name != "PostgreSQL" || level != "middle" {
		t.Fatalf("got %q %q", name, level)
	}
}
