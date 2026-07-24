package dialogue_test

import (
	"context"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/dialogue"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

type stubAgent struct {
	reply string
}

func (s stubAgent) Handle(_ context.Context, req ai.DialogueRequest) (ai.DialogueResponse, error) {
	return ai.DialogueResponse{
		Reply:   s.reply,
		History: append(append([]ai.DialogueTurn{}, req.History...), ai.DialogueTurn{Role: "user", Content: req.Text}),
	}, nil
}

func TestServiceTurnNotConfigured(t *testing.T) {
	t.Parallel()
	var s *dialogue.Service
	_, err := s.Turn(context.Background(), ai.DialogueRequest{UserID: ids.NewUserID(), Text: "hi"})
	if err != dialogue.ErrNotConfigured {
		t.Fatalf("got %v", err)
	}
}

func TestServiceTurn(t *testing.T) {
	t.Parallel()
	s := &dialogue.Service{Agent: stubAgent{reply: "ok"}}
	resp, err := s.Turn(context.Background(), ai.DialogueRequest{
		UserID: ids.NewUserID(), Text: "привет",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Reply != "ok" {
		t.Fatalf("reply=%q", resp.Reply)
	}
}
