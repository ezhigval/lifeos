package composite

import (
	"context"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type stubResolver struct {
	out ai.ResolvedIntent
}

func (s stubResolver) Resolve(context.Context, ai.ResolveInput) (ai.ResolvedIntent, error) {
	return s.out, nil
}

func TestFallbackResolverUsesPrimaryWhenKnown(t *testing.T) {
	t.Parallel()
	r := NewFallbackResolver(
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentTaskCreate, Title: "x", Confidence: 0.9}},
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentProjectCreate, Title: "y", Confidence: 0.9}},
		func() bool { return true },
	)
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentTaskCreate {
		t.Fatalf("got %s, want task.create", got.Type)
	}
}

func TestFallbackResolverUsesFallbackWhenUnknown(t *testing.T) {
	t.Parallel()
	r := NewFallbackResolver(
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentUnknown}},
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentFinanceIncome, Title: "заказ", AmountCents: 500000, Confidence: 0.8}},
		func() bool { return true },
	)
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "получил пятьдесят тысяч"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentFinanceIncome {
		t.Fatalf("got %s, want finance.income", got.Type)
	}
}

func TestFallbackResolverSkipsWhenDisabled(t *testing.T) {
	t.Parallel()
	r := NewFallbackResolver(
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentUnknown}},
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentTaskCreate, Title: "x"}},
		func() bool { return false },
	)
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s, want unknown", got.Type)
	}
}
