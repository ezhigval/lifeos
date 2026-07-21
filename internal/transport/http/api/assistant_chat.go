package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/dialogue"
)

type chatTurnJSON struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolName   string         `json:"tool_name,omitempty"`
	ToolArgs   map[string]any `json:"tool_args,omitempty"`
	ToolResult string         `json:"tool_result,omitempty"`
}

type assistantChatRequest struct {
	Text     string         `json:"text"`
	History  []chatTurnJSON `json:"history"`
	Language string         `json:"language"`
}

type assistantChatResponse struct {
	Reply    string         `json:"reply"`
	Waiting  bool           `json:"waiting"`
	History  []chatTurnJSON `json:"history"`
	ToolsRun []string       `json:"tools_run"`
}

func (rt *Router) assistantChat(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if rt.deps.Dialogue == nil {
		writeError(w, http.StatusNotImplemented, "assistant is not configured")
		return
	}

	var req assistantChatRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}
	lang := strings.TrimSpace(req.Language)
	if lang == "" {
		lang = "ru"
	}

	resp, err := rt.deps.Dialogue.Turn(r.Context(), ai.DialogueRequest{
		UserID:   userID,
		Text:     text,
		History:  fromChatTurns(req.History),
		Language: lang,
	})
	if err != nil {
		if errors.Is(err, dialogue.ErrNotConfigured) {
			writeError(w, http.StatusNotImplemented, "assistant is not configured")
			return
		}
		if rt.deps.Log != nil {
			rt.deps.Log.Warn("assistant chat failed", "error", err)
		}
		writeError(w, http.StatusBadGateway, "assistant unavailable")
		return
	}

	askRounds := 0
	for _, t := range req.History {
		if t.Role == "assistant" && strings.TrimSpace(t.Content) != "" {
			askRounds++
		}
	}
	if resp.Waiting {
		askRounds++
	}
	rt.deps.Dialogue.RecordLearning(r.Context(), userID, resp, askRounds)

	reply := strings.TrimSpace(resp.Reply)
	if reply == "" {
		reply = "Готово."
	}
	tools := resp.ToolsRun
	if tools == nil {
		tools = []string{}
	}
	writeJSON(w, http.StatusOK, assistantChatResponse{
		Reply:    reply,
		Waiting:  resp.Waiting,
		History:  toChatTurns(resp.History),
		ToolsRun: tools,
	})
}

func fromChatTurns(in []chatTurnJSON) []ai.DialogueTurn {
	if len(in) == 0 {
		return nil
	}
	out := make([]ai.DialogueTurn, 0, len(in))
	for _, t := range in {
		role := strings.TrimSpace(t.Role)
		if role == "" {
			continue
		}
		out = append(out, ai.DialogueTurn{
			Role:       role,
			Content:    t.Content,
			ToolName:   t.ToolName,
			ToolArgs:   t.ToolArgs,
			ToolResult: t.ToolResult,
		})
	}
	return out
}

func toChatTurns(in []ai.DialogueTurn) []chatTurnJSON {
	out := make([]chatTurnJSON, 0, len(in))
	for _, t := range in {
		out = append(out, chatTurnJSON{
			Role:       t.Role,
			Content:    t.Content,
			ToolName:   t.ToolName,
			ToolArgs:   t.ToolArgs,
			ToolResult: t.ToolResult,
		})
	}
	return out
}
