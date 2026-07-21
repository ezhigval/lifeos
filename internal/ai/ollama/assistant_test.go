package ollama

import (
	"strings"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
)

func TestFormatSummaryRequest(t *testing.T) {
	t.Parallel()
	got := formatSummaryRequest(ai.SummaryRequest{
		Tasks:    []string{"бег", "звонок"},
		Projects: []string{"500к"},
	})
	if !strings.Contains(got, "бег") || !strings.Contains(got, "500к") {
		t.Fatalf("got %q", got)
	}
}

func TestFormatSummaryRequestEmpty(t *testing.T) {
	t.Parallel()
	got := formatSummaryRequest(ai.SummaryRequest{})
	if !strings.Contains(got, "Задач на сегодня нет") {
		t.Fatalf("got %q", got)
	}
}
