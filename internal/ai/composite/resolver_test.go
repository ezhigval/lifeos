package composite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type stubResolver struct {
	out   ai.ResolvedIntent
	err   error
	delay time.Duration
	calls *int
}

func (s stubResolver) Resolve(ctx context.Context, _ ai.ResolveInput) (ai.ResolvedIntent, error) {
	if s.calls != nil {
		*s.calls++
	}
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return ai.ResolvedIntent{}, ctx.Err()
		}
	}
	if s.err != nil {
		return ai.ResolvedIntent{}, s.err
	}
	return s.out, nil
}

func TestFallbackResolverUsesPrimaryWhenKnown(t *testing.T) {
	t.Parallel()
	fbCalls := 0
	r := NewFallbackResolver(
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentTaskCreate, Title: "x", Confidence: 0.9}},
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentProjectCreate, Title: "y", Confidence: 0.9}, calls: &fbCalls},
		func() bool { return true },
	)
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentTaskCreate {
		t.Fatalf("got %s, want task.create", got.Type)
	}
	if fbCalls != 0 {
		t.Fatalf("fallback called %d times, want 0", fbCalls)
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

func TestFallbackResolverDegradesOnFallbackError(t *testing.T) {
	t.Parallel()
	r := NewFallbackResolver(
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentUnknown}},
		stubResolver{err: errors.New("ollama down")},
		func() bool { return true },
	)
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: " gibberish "})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s, want unknown", got.Type)
	}
}

func TestFallbackResolverDegradesOnFallbackTimeout(t *testing.T) {
	t.Parallel()
	r := NewFallbackResolver(
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentUnknown}},
		stubResolver{
			out:   ai.ResolvedIntent{Type: ai.IntentTaskCreate, Title: "late"},
			delay: 200 * time.Millisecond,
		},
		func() bool { return true },
	).WithTimeout(20 * time.Millisecond)

	start := time.Now()
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "slow llm"})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s, want unknown after timeout", got.Type)
	}
	if elapsed > 150*time.Millisecond {
		t.Fatalf("timeout took too long: %v", elapsed)
	}
}

func TestFallbackResolverDegradesWhenFallbackUnknown(t *testing.T) {
	t.Parallel()
	r := NewFallbackResolver(
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentUnknown, Confidence: 0.1}},
		stubResolver{out: ai.ResolvedIntent{Type: ai.IntentUnknown}},
		func() bool { return true },
	)
	got, err := r.Resolve(context.Background(), ai.ResolveInput{Text: "???"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != ai.IntentUnknown {
		t.Fatalf("got %s", got.Type)
	}
}
