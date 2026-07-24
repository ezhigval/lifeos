package telegram

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/dialogue"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tginfra "github.com/valentinezhov/lifeos/internal/transport/telegram/infra"
)

const (
	payloadAgentHistory = "agent_history"
	payloadAgentAsks    = "agent_ask_rounds"
	maxAgentHistory     = 16
)

// AgentBridge wraps the shared dialogue service for Telegram session history.
type AgentBridge struct {
	*dialogue.Service
}

func (h *MessageHandler) runAgentDialogue(ctx context.Context, userID ids.UserID, text string, sess tginfra.Session) (dispatchResult, error) {
	if h.agent == nil || h.agent.Service == nil || h.agent.Agent == nil {
		intent, err := h.resolver.Resolve(ctx, ai.ResolveInput{Text: text, Language: "ru"})
		if err != nil {
			return dispatchResult{}, err
		}
		return h.dispatchIntent(ctx, userID, intent)
	}

	history := loadAgentHistory(sess.StatePayload)
	askRounds := payloadInt(sess.StatePayload, payloadAgentAsks)

	resp, err := h.agent.Turn(ctx, ai.DialogueRequest{
		UserID:   userID,
		Text:     text,
		History:  history,
		Language: "ru",
	})
	if err != nil {
		h.log.Warn("agent dialogue failed, falling back to intent resolver", "error", err)
		intent, rerr := h.resolver.Resolve(ctx, ai.ResolveInput{Text: text, Language: "ru"})
		if rerr != nil {
			return dispatchResult{}, rerr
		}
		return h.dispatchIntent(ctx, userID, intent)
	}

	payload := h.basePayload(ctx, userID)
	for k, v := range filterAgentKeys(sess.StatePayload) {
		payload[k] = v
	}
	if resp.Waiting {
		askRounds++
		payload[payloadAgentAsks] = askRounds
		payload[payloadAgentHistory] = truncateHistory(resp.History)
		_ = h.sessions.SetState(ctx, userID, tginfra.StateAwaitAgentTurn, payload)
	} else {
		delete(payload, payloadAgentHistory)
		delete(payload, payloadAgentAsks)
		_ = h.sessions.SetState(ctx, userID, tginfra.StateIdle, payload)
	}

	h.agent.RecordLearning(ctx, userID, resp, askRounds)

	reply := strings.TrimSpace(resp.Reply)
	if reply == "" {
		reply = "Готово."
	}
	return dispatchResult{text: reply}, nil
}

func loadAgentHistory(payload map[string]any) []ai.DialogueTurn {
	if payload == nil {
		return nil
	}
	raw, ok := payload[payloadAgentHistory]
	if !ok || raw == nil {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var turns []ai.DialogueTurn
	if err := json.Unmarshal(b, &turns); err != nil {
		return nil
	}
	return turns
}

func truncateHistory(turns []ai.DialogueTurn) []ai.DialogueTurn {
	if len(turns) <= maxAgentHistory {
		return turns
	}
	return turns[len(turns)-maxAgentHistory:]
}

func filterAgentKeys(payload map[string]any) map[string]any {
	out := map[string]any{}
	if payload == nil {
		return out
	}
	for _, k := range []string{payloadAgentHistory, payloadAgentAsks} {
		if v, ok := payload[k]; ok {
			out[k] = v
		}
	}
	return out
}

func payloadInt(payload map[string]any, key string) int {
	if payload == nil {
		return 0
	}
	v, ok := payload[key]
	if !ok {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}
