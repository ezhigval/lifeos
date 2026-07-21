package api_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/ai/dialogue"
	identityapp "github.com/valentinezhov/lifeos/internal/identity/app"
	"github.com/valentinezhov/lifeos/internal/platform/auth"
	"github.com/valentinezhov/lifeos/internal/transport/http/api"
)

type chatStubAgent struct{}

func (chatStubAgent) Handle(_ context.Context, req ai.DialogueRequest) (ai.DialogueResponse, error) {
	hist := append(append([]ai.DialogueTurn{}, req.History...), ai.DialogueTurn{Role: "user", Content: req.Text})
	hist = append(hist, ai.DialogueTurn{Role: "assistant", Content: "понял: " + req.Text})
	return ai.DialogueResponse{Reply: "понял: " + req.Text, History: hist}, nil
}

func TestAssistantChat(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	users := &stubUserRepo{user: env.user}
	tokens, err := auth.NewTokenService(testJWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	rt := api.NewRouter(api.Deps{
		Log:         slog.Default(),
		APIKey:      testAPIKey,
		Tokens:      tokens,
		GetUser:     identityapp.NewGetUserByTelegram(users),
		GetUserByID: identityapp.NewGetUserByID(users),
		Dialogue:    &dialogue.Service{Agent: chatStubAgent{}},
	})
	r := chi.NewRouter()
	rt.Mount(r)

	token := issueToken(t, testEnv{user: env.user, router: r})
	rec := doJSON(t, r, http.MethodPost, "/api/v1/assistant/chat",
		map[string]string{"Authorization": "Bearer " + token},
		map[string]any{"text": "привет"},
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out struct {
		Reply   string `json:"reply"`
		Waiting bool   `json:"waiting"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Reply != "понял: привет" || out.Waiting {
		t.Fatalf("out=%+v", out)
	}
}

func TestAssistantChatNotConfigured(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	token := issueToken(t, env)
	rec := doJSON(t, env.router, http.MethodPost, "/api/v1/assistant/chat",
		map[string]string{"Authorization": "Bearer " + token},
		map[string]any{"text": "hi"},
	)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status=%d want 501 body=%s", rec.Code, rec.Body.String())
	}
}
