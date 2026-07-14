package telegram

import (
	"encoding/json"
	"testing"
)

func TestReplyKeyboardMarkupShape(t *testing.T) {
	t.Parallel()
	m := replyKeyboardMarkup([][]string{{"A", "B"}, {"C"}})
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	kb, ok := decoded["keyboard"].([]any)
	if !ok || len(kb) != 2 {
		t.Fatalf("keyboard rows: %#v", decoded["keyboard"])
	}
	row0, ok := kb[0].([]any)
	if !ok || len(row0) != 2 {
		t.Fatalf("row0: %#v", kb[0])
	}
	btn, ok := row0[0].(map[string]any)
	if !ok || btn["text"] != "A" {
		t.Fatalf("button should be {text: A}, got %#v", row0[0])
	}
	if decoded["is_persistent"] != true || decoded["resize_keyboard"] != true {
		t.Fatalf("flags: %#v", decoded)
	}
}

func TestReplyKeyboardInstalled(t *testing.T) {
	t.Parallel()
	if replyKeyboardInstalled(nil) {
		t.Fatal("nil")
	}
	if replyKeyboardInstalled(map[string]any{replyKBSetKey: true}) {
		t.Fatal("missing version must reinstall")
	}
	if !replyKeyboardInstalled(map[string]any{
		replyKBSetKey:     true,
		replyKBVersionKey: float64(replyKBVersion),
	}) {
		t.Fatal("expected installed")
	}
	if replyKeyboardInstalled(map[string]any{
		replyKBSetKey:     true,
		replyKBVersionKey: float64(1),
	}) {
		t.Fatal("old version must reinstall")
	}
}
