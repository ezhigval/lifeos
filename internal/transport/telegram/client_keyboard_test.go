package telegram

import (
	"encoding/json"
	"testing"
)

func TestReplyKeyboardMarkupShape(t *testing.T) {
	t.Parallel()
	m := replyKeyboardMarkup([][]ReplyButton{
		{{Text: "A"}, {Text: "B"}},
		{{Text: "C", WebApp: "https://example.com/app/"}},
	})
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
	row1, ok := kb[1].([]any)
	if !ok || len(row1) != 1 {
		t.Fatalf("row1: %#v", kb[1])
	}
	webBtn, ok := row1[0].(map[string]any)
	if !ok {
		t.Fatalf("web button: %#v", row1[0])
	}
	webApp, ok := webBtn["web_app"].(map[string]any)
	if !ok || webApp["url"] != "https://example.com/app/" {
		t.Fatalf("web_app=%#v", webBtn["web_app"])
	}
	if decoded["is_persistent"] != true || decoded["resize_keyboard"] != true {
		t.Fatalf("flags: %#v", decoded)
	}
}

func TestReplyKeyboardInstalled(t *testing.T) {
	t.Parallel()
	if replyKeyboardInstalled(nil, false) {
		t.Fatal("nil")
	}
	if replyKeyboardInstalled(map[string]any{replyKBSetKey: true}, false) {
		t.Fatal("missing version must reinstall")
	}
	if !replyKeyboardInstalled(map[string]any{
		replyKBSetKey:     true,
		replyKBVersionKey: float64(replyKBVersion),
		replyKBMiniAppKey: false,
	}, false) {
		t.Fatal("expected installed without mini app")
	}
	if replyKeyboardInstalled(map[string]any{
		replyKBSetKey:     true,
		replyKBVersionKey: float64(replyKBVersion),
		replyKBMiniAppKey: false,
	}, true) {
		t.Fatal("mini app presence mismatch must reinstall")
	}
	if !replyKeyboardInstalled(map[string]any{
		replyKBSetKey:     true,
		replyKBVersionKey: float64(replyKBVersion),
		replyKBMiniAppKey: true,
	}, true) {
		t.Fatal("expected installed with mini app")
	}
	if replyKeyboardInstalled(map[string]any{
		replyKBSetKey:     true,
		replyKBVersionKey: float64(1),
		replyKBMiniAppKey: false,
	}, false) {
		t.Fatal("old version must reinstall")
	}
}

func TestReplyKeyboardHasMiniApp(t *testing.T) {
	t.Parallel()
	if replyKeyboardHasMiniApp(MainReplyKeyboard("")) {
		t.Fatal("empty URL must omit web_app")
	}
	if !replyKeyboardHasMiniApp(MainReplyKeyboard("https://example.com/app/")) {
		t.Fatal("URL must add web_app row")
	}
}
