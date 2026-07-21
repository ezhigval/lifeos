package openaiapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/reviewsafe"
)

type Assistant struct {
	client *Client
}

func NewAssistant(client *Client) *Assistant {
	return &Assistant{client: client}
}

func (a *Assistant) Summarize(ctx context.Context, req ai.SummaryRequest) (string, error) {
	if a.client == nil {
		return "", fmt.Errorf("openaiapi client is nil")
	}
	text, err := a.client.Chat(ctx, reviewSystemPrompt, formatSummaryRequest(req))
	if err != nil {
		return "", err
	}
	text = reviewsafe.SanitizeHTML(text)
	if text == "" {
		return "", fmt.Errorf("empty summary")
	}
	return text, nil
}

func formatSummaryRequest(req ai.SummaryRequest) string {
	var b strings.Builder
	b.WriteString("Данные для обзора:\n")
	if len(req.Tasks) == 0 {
		b.WriteString("Задач на сегодня нет.\n")
	} else {
		b.WriteString("Задачи на сегодня:\n")
		for i, t := range req.Tasks {
			fmt.Fprintf(&b, "%d. %s\n", i+1, t)
		}
	}
	if len(req.Projects) == 0 {
		b.WriteString("Активных проектов нет.\n")
	} else {
		b.WriteString("Активные проекты:\n")
		for _, p := range req.Projects {
			fmt.Fprintf(&b, "- %s\n", p)
		}
	}
	return strings.TrimSpace(b.String())
}
