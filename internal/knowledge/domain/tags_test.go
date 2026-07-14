package domain_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/knowledge/domain"
)

func TestExtractHashtags(t *testing.T) {
	t.Parallel()
	body, tags := domain.ExtractHashtags("#work идея для #Jarvis")
	if body != "идея для" {
		t.Fatalf("body=%q", body)
	}
	if len(tags) != 2 {
		t.Fatalf("tags=%v", tags)
	}
	tagSet := map[string]bool{tags[0]: true, tags[1]: true}
	if !tagSet["work"] || !tagSet["jarvis"] {
		t.Fatalf("tags=%v", tags)
	}
}
