package rulebased_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
)

func TestResolverGoldenFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "intents.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var cases []struct {
		Input       string        `json:"input"`
		Intent      ai.IntentType `json:"intent"`
		Title       string        `json:"title"`
		Target      string        `json:"target"`
		AmountCents int64         `json:"amount_cents"`
		Unit        string        `json:"unit"`
		Hour        int           `json:"hour"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}

	r := rulebased.NewResolver()
	for _, tc := range cases {
		got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: tc.Input})
		if err != nil {
			t.Fatalf("Resolve(%q): %v", tc.Input, err)
		}
		if got.Type != tc.Intent {
			t.Fatalf("Resolve(%q) intent = %s, want %s", tc.Input, got.Type, tc.Intent)
		}
		if tc.Title != "" && got.Title != tc.Title {
			t.Fatalf("Resolve(%q) title = %q, want %q", tc.Input, got.Title, tc.Title)
		}
		if tc.Target != "" && got.Target != tc.Target {
			t.Fatalf("Resolve(%q) target = %q, want %q", tc.Input, got.Target, tc.Target)
		}
		if tc.AmountCents != 0 && got.AmountCents != tc.AmountCents {
			t.Fatalf("Resolve(%q) amount = %d, want %d", tc.Input, got.AmountCents, tc.AmountCents)
		}
		if tc.Unit != "" && got.Unit != tc.Unit {
			t.Fatalf("Resolve(%q) unit = %q, want %q", tc.Input, got.Unit, tc.Unit)
		}
		if tc.Hour != 0 && got.Hour != tc.Hour {
			t.Fatalf("Resolve(%q) hour = %d, want %d", tc.Input, got.Hour, tc.Hour)
		}
	}
}
