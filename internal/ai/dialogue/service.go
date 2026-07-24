package dialogue

import (
	"context"
	"strings"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/learning"
	learningapp "github.com/valentinezhov/lifeos/internal/learning/app"
	memoryapp "github.com/valentinezhov/lifeos/internal/memory/app"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

// Service is the transport-agnostic conversational turn runner
// (Telegram + Mini App HTTP).
type Service struct {
	Agent        ai.ConversationalAgent
	ListMemories *memoryapp.ListMemories
	Privacy      memoryapp.PrivacyStore
	RecordLearn  *learningapp.RecordEvent
	LearningSalt string
	ModelName    string
}

// Turn loads personal memory (if enabled), runs the agent, and optionally
// records anonymous learning. Callers own multi-turn history.
func (s *Service) Turn(ctx context.Context, req ai.DialogueRequest) (ai.DialogueResponse, error) {
	if s == nil || s.Agent == nil {
		return ai.DialogueResponse{}, ErrNotConfigured
	}

	memoryOn := true
	if s.Privacy != nil {
		if flags, err := s.Privacy.GetPrivacyFlags(ctx, req.UserID); err == nil {
			memoryOn = flags.MemoryEnabled
		}
	}
	if memoryOn && s.ListMemories != nil && len(req.Memories) == 0 {
		if items, err := s.ListMemories.Execute(ctx, req.UserID, 12); err == nil {
			for _, m := range items {
				req.Memories = append(req.Memories, ai.MemorySnippet{
					Kind: string(m.Kind), Key: m.Key, Value: m.Value,
				})
			}
		}
	}

	resp, err := s.Agent.Handle(ctx, req)
	if err != nil {
		return ai.DialogueResponse{}, err
	}
	return resp, nil
}

// RecordLearning writes an anonymized dialogue_turn event when the user opted in.
func (s *Service) RecordLearning(ctx context.Context, userID ids.UserID, resp ai.DialogueResponse, askRounds int) {
	if s == nil || s.RecordLearn == nil || s.Privacy == nil {
		return
	}
	flags, err := s.Privacy.GetPrivacyFlags(ctx, userID)
	if err != nil || !flags.LearningOptIn {
		return
	}
	success := len(resp.ToolsRun) > 0 && !resp.Waiting
	meta := map[string]any{
		"tools_count": len(resp.ToolsRun),
		"waiting":     resp.Waiting,
	}
	if len(resp.ToolsRun) > 0 {
		meta["tools"] = resp.ToolsRun
	}
	_ = s.RecordLearn.Execute(ctx, learningapp.EventInput{
		AnonSubject:  learning.AnonSubject(userID, s.LearningSalt),
		Type:         "dialogue_turn",
		ToolOrIntent: strings.Join(resp.ToolsRun, ","),
		Success:      &success,
		AskRounds:    askRounds,
		Model:        s.ModelName,
		Meta:         meta,
	})
}
