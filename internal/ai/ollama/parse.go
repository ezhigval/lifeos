package ollama

import (
	"encoding/json"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/rulebased"
)

type llmPayload struct {
	Intent       string   `json:"intent"`
	Title        string   `json:"title"`
	Target       string   `json:"target"`
	Unit         string   `json:"unit"`
	AmountRubles *float64 `json:"amount_rubles"`
	Hour         *int     `json:"hour"`
	Minute       *int     `json:"minute"`
	Confidence   float64  `json:"confidence"`
}

func parseResponse(raw string) (ai.ResolvedIntent, error) {
	raw = strings.TrimSpace(raw)
	var p llmPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return ai.ResolvedIntent{Type: ai.IntentUnknown, Confidence: 0}, nil
	}
	intent := ai.IntentType(strings.TrimSpace(p.Intent))
	if intent == "" {
		intent = ai.IntentUnknown
	}
	out := ai.ResolvedIntent{
		Type:       intent,
		Title:      strings.TrimSpace(p.Title),
		Target:     strings.TrimSpace(p.Target),
		Unit:       strings.TrimSpace(p.Unit),
		Confidence: p.Confidence,
		Currency:   "RUB",
	}
	if p.Hour != nil {
		out.Hour = *p.Hour
	}
	if p.Minute != nil {
		out.Minute = *p.Minute
	}
	if p.AmountRubles != nil && *p.AmountRubles > 0 {
		out.AmountCents = int64(*p.AmountRubles * 100)
	}
	if out.Confidence <= 0 {
		out.Confidence = 0.7
	}
	if out.Type == ai.IntentUnknown || out.Confidence < 0.4 {
		return ai.ResolvedIntent{Type: ai.IntentUnknown, Confidence: 0}, nil
	}
	if out.AmountCents == 0 {
		switch out.Type {
		case ai.IntentFinanceIncome, ai.IntentFinanceExpense:
			if cents, err := rulebased.ParseRublesAmount(out.Title); err == nil {
				out.AmountCents = cents
			}
		case ai.IntentFinanceCreateDebt:
			if cents, err := rulebased.ParseRublesAmount(out.Target); err == nil {
				out.AmountCents = cents
			}
		case ai.IntentFinancePayDebt:
			if cents, err := rulebased.ParseRublesAmount(out.Title); err == nil {
				out.AmountCents = cents
			}
			if out.AmountCents == 0 {
				if cents, err := rulebased.ParseRublesAmount(out.Target); err == nil {
					out.AmountCents = cents
				}
			}
		}
	}
	return out, nil
}
