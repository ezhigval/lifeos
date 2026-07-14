package rulebased_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
)

func TestParseWeightKg(t *testing.T) {
	t.Parallel()
	w, err := rulebased.ParseWeightKg("78,5 кг")
	if err != nil || w != 78.5 {
		t.Fatalf("w=%v err=%v", w, err)
	}
}
