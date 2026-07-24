package openaiapi

import (
	"context"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type Resolver struct {
	client *Client
}

func NewResolver(client *Client) *Resolver {
	return &Resolver{client: client}
}

func (r *Resolver) Resolve(ctx context.Context, input ai.ResolveInput) (ai.ResolvedIntent, error) {
	if r.client == nil {
		return ai.ResolvedIntent{Type: ai.IntentUnknown, Confidence: 0}, nil
	}
	raw, err := r.client.ChatJSON(ctx, systemPrompt, input.Text)
	if err != nil {
		return ai.ResolvedIntent{Type: ai.IntentUnknown, Confidence: 0}, nil
	}
	return parseResponse(raw)
}
