package agent

import (
	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/ollama"
	"github.com/valentinezhov/lifeos/internal/ai/openaiapi"
)

// Compile-time checks: existing LLM clients already satisfy ai.ChatModel.
var (
	_ ai.ChatModel = (*ollama.Client)(nil)
	_ ai.ChatModel = (*openaiapi.Client)(nil)
)
