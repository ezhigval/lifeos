package composite

import (
	"context"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai"
)

// DefaultFallbackTimeout is the max wait for the optional LLM branch.
// Rule-based primary never blocks on this; only Ollama/network does.
const DefaultFallbackTimeout = 8 * time.Second

// FallbackResolver tries the rule-based primary first. When the primary returns
// IntentUnknown and LLM is enabled, it asks the fallback (Ollama). Any fallback
// error, timeout, or unknown result is ignored — the primary unknown stands.
//
// Production order (wired in cmd/lifeos/cmd/resolver.go):
//  1. rulebased for known intents (fast, offline)
//  2. optional LLM only for unknown intents
//  3. degrade (logged when OnDegrade set) when LLM is down or slow
type FallbackResolver struct {
	primary   ai.IntentResolver
	fallback  ai.IntentResolver
	enabled   func() bool
	timeout   time.Duration
	onDegrade func(reason string, err error)
}

func NewFallbackResolver(primary, fallback ai.IntentResolver, enabled func() bool) *FallbackResolver {
	return &FallbackResolver{
		primary:  primary,
		fallback: fallback,
		enabled:  enabled,
		timeout:  DefaultFallbackTimeout,
	}
}

// WithTimeout overrides the LLM fallback wait (tests / tuning).
func (r *FallbackResolver) WithTimeout(d time.Duration) *FallbackResolver {
	if d > 0 {
		r.timeout = d
	}
	return r
}

// WithOnDegrade registers a callback when the LLM branch fails or returns unknown.
func (r *FallbackResolver) WithOnDegrade(fn func(reason string, err error)) *FallbackResolver {
	r.onDegrade = fn
	return r
}

func (r *FallbackResolver) Resolve(ctx context.Context, input ai.ResolveInput) (ai.ResolvedIntent, error) {
	got, err := r.primary.Resolve(ctx, input)
	if err != nil {
		return ai.ResolvedIntent{}, err
	}
	if got.Type != ai.IntentUnknown {
		return got, nil
	}
	if r.fallback == nil || (r.enabled != nil && !r.enabled()) {
		return got, nil
	}

	fbCtx := ctx
	cancel := func() {}
	if r.timeout > 0 {
		fbCtx, cancel = context.WithTimeout(ctx, r.timeout)
	}
	defer cancel()

	fb, err := r.fallback.Resolve(fbCtx, input)
	if err != nil {
		if r.onDegrade != nil {
			r.onDegrade("llm_error", err)
		}
		return got, nil
	}
	if fb.Type == ai.IntentUnknown {
		if r.onDegrade != nil {
			r.onDegrade("llm_unknown", nil)
		}
		return got, nil
	}
	return fb, nil
}
