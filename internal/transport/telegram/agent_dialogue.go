package telegram

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/learning"
	learningapp "github.com/valentinezhov/lifeos/internal/learning/app"
	memoryapp "github.com/valentinezhov/lifeos/internal/memory/app"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
	tginfra "github.com/valentinezhov/lifeos/internal/transport/telegram/infra"
)

const (
	payloadAgentHistory = "agent_history"
	payloadAgentAsks    = "agent_ask_rounds"
	maxAgentHistory     = 16
)

// AgentBridge loads memory, runs the conversational agent, and records anon learning.
type AgentBridge struct {
	Agent        ai.ConversationalAgent
	ListMemories *memoryapp.ListMemories
	Privacy      memoryapp.PrivacyStore
	RecordLearn  *learningapp.RecordEvent
	LearningSalt string
	ModelName    string
}

func (h *MessageHandler) runAgentDialogue(ctx context.Context, userID ids.UserID, text string, sess tginfra.Session) (dispatchResult, error) {
	if h.agent == nil || h.agent.Agent == nil {
		intent, err := h.resolver.Resolve(ctx, ai.ResolveInput{Text: text, Language: "ru"})
		if err != nil {
			return dispatchResult{}, err
		}
		return h.dispatchIntent(ctx, userID, intent)
	}

	history := loadAgentHistory(sess.StatePayload)
	askRounds := payloadInt(sess.StatePayload, payloadAgentAsks)

	var memories []ai.MemorySnippet
	memoryOn := true
	if h.agent.Privacy != nil {
		if flags, err := h.agent.Privacy.GetPrivacyFlags(ctx, userID); err == nil {
			memoryOn = flags.MemoryEnabled
			// learning opt-in checked after turn
			_ = flags.LearningOptIn
		}
	}
	if memoryOn && h.agent.ListMemories != nil {
		if items, err := h.agent.ListMemories.Execute(ctx, userID, 12); err == nil {
			for _, m := range items {
				memories = append(memories, ai.MemorySnippet{
					Kind: string(m.Kind), Key: m.Key, Value: m.Value,
				})
			}
		}
	}

	resp, err := h.agent.Agent.Handle(ctx, ai.DialogueRequest{
		UserID:   userID,
		Text:     text,
		History:  history,
		Memories: memories,
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

	h.maybeRecordLearning(ctx, userID, resp, askRounds)

	reply := strings.TrimSpace(resp.Reply)
	if reply == "" {
		reply = "Готово."
	}
	return dispatchResult{text: reply}, nil
}

func (h *MessageHandler) maybeRecordLearning(ctx context.Context, userID ids.UserID, resp ai.DialogueResponse, askRounds int) {
	if h.agent == nil || h.agent.RecordLearn == nil || h.agent.Privacy == nil {
		return
	}
	flags, err := h.agent.Privacy.GetPrivacyFlags(ctx, userID)
	if err != nil || !flags.LearningOptIn {
		return
	}
	success := len(resp.ToolsRun) > 0 && !resp.Waiting
	meta := map[string]any{
		"tools_count": len(resp.ToolsRun),
		"waiting":     resp.Waiting,
	}
	// Coarse tool names only — no args, no user text.
	if len(resp.ToolsRun) > 0 {
		meta["tools"] = resp.ToolsRun
	}
	_ = h.agent.RecordLearn.Execute(ctx, learningapp.EventInput{
		AnonSubject:  learning.AnonSubject(userID, h.agent.LearningSalt),
		Type:         "dialogue_turn",
		ToolOrIntent: strings.Join(resp.ToolsRun, ","),
		Success:      &success,
		AskRounds:    askRounds,
		Model:        h.agent.ModelName,
		Meta:         meta,
	})
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
