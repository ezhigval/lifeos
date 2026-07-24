package agent

import (
	"context"
	"fmt"

	"github.com/valentinezhov/lifeos/internal/ai"
)

const maxToolRounds = 4

const softRephraseReply = "Не совсем понял запрос. Переформулируйте, пожалуйста, чуть конкретнее."

// Agent is a multi-turn conversational orchestrator over a ChatModel and tool registry.
type Agent struct {
	model ai.ChatModel
	tools *Registry
}

type Option func(*Agent)

func New(model ai.ChatModel, tools *Registry, opts ...Option) *Agent {
	a := &Agent{model: model, tools: tools}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Handle implements ai.ConversationalAgent.
func (a *Agent) Handle(ctx context.Context, req ai.DialogueRequest) (ai.DialogueResponse, error) {
	if a == nil || a.model == nil {
		return ai.DialogueResponse{}, fmt.Errorf("agent: model is required")
	}

	history := append([]ai.DialogueTurn(nil), req.History...)
	history = append(history, ai.DialogueTurn{Role: "user", Content: req.Text})

	system := buildSystemPrompt(a.tools, req.Memories, req.Language)
	var toolsRun []string

	for round := 0; round < maxToolRounds; round++ {
		userBlob := serializeUserBlob(req, history)
		raw, err := a.model.ChatJSON(ctx, system, userBlob)
		if err != nil {
			return ai.DialogueResponse{}, fmt.Errorf("agent chat: %w", err)
		}

		action, err := parseAction(raw)
		if err != nil {
			history = append(history, ai.DialogueTurn{Role: "assistant", Content: softRephraseReply})
			return ai.DialogueResponse{
				Reply:    softRephraseReply,
				Waiting:  false,
				History:  history,
				ToolsRun: toolsRun,
			}, nil
		}

		switch action.Type {
		case actionAsk:
			history = append(history, ai.DialogueTurn{Role: "assistant", Content: action.Text})
			return ai.DialogueResponse{
				Reply:    action.Text,
				Waiting:  true,
				History:  history,
				ToolsRun: toolsRun,
			}, nil

		case actionReply:
			history = append(history, ai.DialogueTurn{Role: "assistant", Content: action.Text})
			return ai.DialogueResponse{
				Reply:    action.Text,
				Waiting:  false,
				History:  history,
				ToolsRun: toolsRun,
			}, nil

		case actionTool:
			result, callErr := a.runTool(ctx, req, action)
			toolsRun = append(toolsRun, action.Tool)
			toolTurn := ai.DialogueTurn{
				Role:       "tool",
				ToolName:   action.Tool,
				ToolArgs:   action.Args,
				ToolResult: result,
			}
			if callErr != nil {
				toolTurn.ToolResult = fmt.Sprintf("error: %v", callErr)
			}
			if action.Text != "" {
				toolTurn.Content = action.Text
			}
			history = append(history, toolTurn)
			// Continue loop so the model can reply with tool results.
			continue
		}
	}

	// Exhausted tool rounds without a final ask/reply.
	fallback := "Сделал шаги по инструментам, но не успел сформулировать ответ. Уточните, что ещё нужно?"
	history = append(history, ai.DialogueTurn{Role: "assistant", Content: fallback})
	return ai.DialogueResponse{
		Reply:    fallback,
		Waiting:  true,
		History:  history,
		ToolsRun: toolsRun,
	}, nil
}

func (a *Agent) runTool(ctx context.Context, req ai.DialogueRequest, action modelAction) (string, error) {
	if a.tools == nil {
		return "", fmt.Errorf("no tools registered")
	}
	return a.tools.Call(ctx, req.UserID, action.Tool, action.Args)
}

var _ ai.ConversationalAgent = (*Agent)(nil)
