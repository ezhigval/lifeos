package agent

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/valentinezhov/lifeos/internal/ai"
	"github.com/valentinezhov/lifeos/internal/platform/ids"
)

func TestParseAction_JSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		raw     string
		wantTyp actionType
		wantTxt string
		wantTool string
		wantErr bool
	}{
		{
			name:    "ask",
			raw:     `{"type":"ask","text":"Какой срок?"}`,
			wantTyp: actionAsk,
			wantTxt: "Какой срок?",
		},
		{
			name:    "reply",
			raw:     `{"type":"reply","text":"Готово"}`,
			wantTyp: actionReply,
			wantTxt: "Готово",
		},
		{
			name:     "tool",
			raw:      `{"type":"tool","tool":"task.create","args":{"title":"Купить молоко"}}`,
			wantTyp:  actionTool,
			wantTool: "task.create",
		},
		{
			name:    "fenced",
			raw:     "```json\n{\"type\":\"reply\",\"text\":\"ok\"}\n```",
			wantTyp: actionReply,
			wantTxt: "ok",
		},
		{
			name:    "invalid",
			raw:     `not json`,
			wantErr: true,
		},
		{
			name:    "ask without text",
			raw:     `{"type":"ask"}`,
			wantErr: true,
		},
		{
			name:    "tool without name",
			raw:     `{"type":"tool","args":{}}`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseAction(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAction: %v", err)
			}
			if got.Type != tc.wantTyp {
				t.Fatalf("type=%q want %q", got.Type, tc.wantTyp)
			}
			if tc.wantTxt != "" && got.Text != tc.wantTxt {
				t.Fatalf("text=%q want %q", got.Text, tc.wantTxt)
			}
			if tc.wantTool != "" && got.Tool != tc.wantTool {
				t.Fatalf("tool=%q want %q", got.Tool, tc.wantTool)
			}
			if got.Type == actionTool && got.Args == nil {
				t.Fatal("tool args should be non-nil map")
			}
		})
	}
}

type scriptedModel struct {
	responses []string
	calls     atomic.Int32
}

func (m *scriptedModel) Chat(ctx context.Context, systemPrompt, userText string) (string, error) {
	return m.ChatJSON(ctx, systemPrompt, userText)
}

func (m *scriptedModel) ChatJSON(ctx context.Context, systemPrompt, userText string) (string, error) {
	i := int(m.calls.Add(1) - 1)
	if i >= len(m.responses) {
		return `{"type":"reply","text":"fallback"}`, nil
	}
	return m.responses[i], nil
}

func TestAgent_ToolThenReply(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	var called atomic.Bool
	var gotTitle string
	reg.Register(ToolTaskCreate, "Create a task", `{"title":"string","due":"string?","priority":"string?"}`,
		func(ctx context.Context, userID ids.UserID, args map[string]any) (string, error) {
			called.Store(true)
			if title, ok := args["title"].(string); ok {
				gotTitle = title
			}
			return "created: " + gotTitle, nil
		},
	)

	toolPayload, _ := json.Marshal(map[string]any{
		"type": "tool",
		"tool": ToolTaskCreate,
		"args": map[string]any{"title": "Купить молоко"},
	})
	model := &scriptedModel{responses: []string{
		string(toolPayload),
		`{"type":"reply","text":"Задача создана"}`,
	}}

	a := New(model, reg)
	resp, err := a.Handle(context.Background(), ai.DialogueRequest{
		UserID: ids.NewUserID(),
		Text:   "добавь задачу купить молоко",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !called.Load() {
		t.Fatal("expected task.create handler to be called")
	}
	if gotTitle != "Купить молоко" {
		t.Fatalf("title=%q", gotTitle)
	}
	if resp.Reply != "Задача создана" {
		t.Fatalf("reply=%q", resp.Reply)
	}
	if resp.Waiting {
		t.Fatal("Waiting should be false for reply")
	}
	if len(resp.ToolsRun) != 1 || resp.ToolsRun[0] != ToolTaskCreate {
		t.Fatalf("ToolsRun=%v", resp.ToolsRun)
	}
	if model.calls.Load() != 2 {
		t.Fatalf("expected 2 model calls, got %d", model.calls.Load())
	}
	foundTool := false
	for _, turn := range resp.History {
		if turn.Role == "tool" && turn.ToolName == ToolTaskCreate {
			foundTool = true
			if turn.ToolResult != "created: Купить молоко" {
				t.Fatalf("tool result=%q", turn.ToolResult)
			}
		}
	}
	if !foundTool {
		t.Fatal("history missing tool turn")
	}
}

func TestAgent_AskSetsWaiting(t *testing.T) {
	t.Parallel()

	model := &scriptedModel{responses: []string{
		`{"type":"ask","text":"На какое время поставить напоминание?"}`,
	}}
	a := New(model, NewRegistry())
	resp, err := a.Handle(context.Background(), ai.DialogueRequest{
		UserID: ids.NewUserID(),
		Text:   "напомни мне",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !resp.Waiting {
		t.Fatal("Waiting should be true for ask")
	}
	if resp.Reply != "На какое время поставить напоминание?" {
		t.Fatalf("reply=%q", resp.Reply)
	}
	if len(resp.ToolsRun) != 0 {
		t.Fatalf("ToolsRun should be empty, got %v", resp.ToolsRun)
	}
}

func TestAgent_ParseErrorSoftReply(t *testing.T) {
	t.Parallel()

	model := &scriptedModel{responses: []string{`totally broken`}}
	a := New(model, NewRegistry())
	resp, err := a.Handle(context.Background(), ai.DialogueRequest{
		UserID: ids.NewUserID(),
		Text:   "привет",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp.Waiting {
		t.Fatal("Waiting should be false on soft rephrase")
	}
	if resp.Reply != softRephraseReply {
		t.Fatalf("reply=%q", resp.Reply)
	}
}
