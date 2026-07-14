package templateassistant

import (
	"context"
	"fmt"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
)

type Assistant struct{}

func New() *Assistant {
	return &Assistant{}
}

func (a *Assistant) Summarize(_ context.Context, req ai.SummaryRequest) (string, error) {
	var b strings.Builder
	b.WriteString("📋 Обзор\n")
	if len(req.Tasks) == 0 {
		b.WriteString("Задач нет.\n")
	} else {
		b.WriteString("Задачи:\n")
		for i, t := range req.Tasks {
			fmt.Fprintf(&b, "%d. %s\n", i+1, t)
		}
	}
	if len(req.Projects) > 0 {
		b.WriteString("Проекты:\n")
		for _, p := range req.Projects {
			fmt.Fprintf(&b, "• %s\n", p)
		}
	}
	return strings.TrimSpace(b.String()), nil
}
