package api

import "testing"

func TestFromChatTurnsDropsClientToolResults(t *testing.T) {
	t.Parallel()
	got := fromChatTurns([]chatTurnJSON{
		{Role: "user", Content: "привет"},
		{Role: "tool", ToolName: "task_create", ToolResult: "создано"},
		{Role: "assistant", Content: "готово"},
		{Role: "  "},
	})
	if len(got) != 2 {
		t.Fatalf("len %d", len(got))
	}
	if got[0].Role != "user" || got[1].Role != "assistant" {
		t.Fatalf("%+v", got)
	}
}
