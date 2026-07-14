package composite

import (
	"context"

	"github.com/valentinezhov/lifeos/internal/ai"
)

// FallbackResolver tries primary first; on unknown intent it delegates to fallback.
type FallbackResolver struct {
	primary  ai.IntentResolver
	fallback ai.IntentResolver
	enabled  func() bool
}

func NewFallbackResolver(primary, fallback ai.IntentResolver, enabled func() bool) *FallbackResolver {
	return &FallbackResolver{primary: primary, fallback: fallback, enabled: enabled}
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
	fb, err := r.fallback.Resolve(ctx, input)
	if err != nil {
		return got, nil
	}
	if fb.Type == ai.IntentUnknown {
		return got, nil
	}
	return fb, nil
}
