package ollama

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
)

func TestParseResponseTaskCreate(t *testing.T) {
	t.Parallel()
	raw := `{"intent":"task.create","title":"купить молоко","confidence":0.92}`
	got, err := parseResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentTaskCreate || got.Title != "купить молоко" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseResponseFinanceIncome(t *testing.T) {
	t.Parallel()
	raw := `{"intent":"finance.income","title":"фриланс","amount_rubles":50000,"confidence":0.88}`
	got, err := parseResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentFinanceIncome || got.AmountCents != 5_000_000 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseResponseUnknownLowConfidence(t *testing.T) {
	t.Parallel()
	raw := `{"intent":"task.create","title":"x","confidence":0.2}`
	got, err := parseResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s", got.Type)
	}
}

func TestParseResponseInvalidJSON(t *testing.T) {
	t.Parallel()
	got, err := parseResponse("not json")
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s", got.Type)
	}
}
