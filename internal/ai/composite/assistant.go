package composite

import (
	"context"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type FallbackAssistant struct {
	primary  ai.Assistant
	fallback ai.Assistant
}

func NewFallbackAssistant(primary, fallback ai.Assistant) *FallbackAssistant {
	return &FallbackAssistant{primary: primary, fallback: fallback}
}

func (a *FallbackAssistant) Summarize(ctx context.Context, req ai.SummaryRequest) (string, error) {
	if a.primary != nil {
		text, err := a.primary.Summarize(ctx, req)
		if err == nil && strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text), nil
		}
	}
	if a.fallback == nil {
		return "", nil
	}
	return a.fallback.Summarize(ctx, req)
}
