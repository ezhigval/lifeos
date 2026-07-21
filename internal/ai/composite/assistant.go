package composite

import (
	"context"
	"strings"
	"time"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/reviewsafe"
)

// DefaultAssistantTimeout is the max wait for the optional LLM summarizer.
const DefaultAssistantTimeout = 15 * time.Second

// FallbackAssistant prefers the LLM primary when enabled; on error, empty,
// unsafe, or timeout it degrades to the template assistant.
type FallbackAssistant struct {
	primary   ai.Assistant
	fallback  ai.Assistant
	timeout   time.Duration
	onDegrade func(reason string, err error)
}

func NewFallbackAssistant(primary, fallback ai.Assistant) *FallbackAssistant {
	return &FallbackAssistant{
		primary:  primary,
		fallback: fallback,
		timeout:  DefaultAssistantTimeout,
	}
}

// WithTimeout overrides the LLM wait (tests / tuning).
func (a *FallbackAssistant) WithTimeout(d time.Duration) *FallbackAssistant {
	if d > 0 {
		a.timeout = d
	}
	return a
}

// WithOnDegrade registers a callback when the LLM summarizer is skipped.
func (a *FallbackAssistant) WithOnDegrade(fn func(reason string, err error)) *FallbackAssistant {
	a.onDegrade = fn
	return a
}

func (a *FallbackAssistant) Summarize(ctx context.Context, req ai.SummaryRequest) (string, error) {
	if a.primary != nil {
		pctx := ctx
		cancel := func() {}
		if a.timeout > 0 {
			pctx, cancel = context.WithTimeout(ctx, a.timeout)
		}
		text, err := a.primary.Summarize(pctx, req)
		cancel()
		if err == nil {
			safe := reviewsafe.SanitizeHTML(text)
			if strings.TrimSpace(safe) != "" {
				return safe, nil
			}
			if a.onDegrade != nil {
				a.onDegrade("empty_or_unsafe", nil)
			}
		} else if a.onDegrade != nil {
			a.onDegrade("llm_error", err)
		}
	}
	if a.fallback == nil {
		return "", nil
	}
	// Template path already EscapePlain's user titles; do not re-sanitize.
	return a.fallback.Summarize(ctx, req)
}
