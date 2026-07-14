package domain_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/career/domain"
)

func TestParseContactLine(t *testing.T) {
	t.Parallel()
	name, company, role := domain.ParseContactLine("Иван Петров — Яндекс")
	if name != "Иван Петров" || company != "Яндекс" || role != "" {
		t.Fatalf("got %q %q %q", name, company, role)
	}
	name, company, role = domain.ParseContactLine("Анна / Sr Go @ Acme")
	if name != "Анна" || company != "Acme" || role != "Sr Go" {
		t.Fatalf("got %q %q %q", name, company, role)
	}
}
