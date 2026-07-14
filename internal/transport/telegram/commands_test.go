package telegram

import "testing"

func TestNormalizeBotCommand(t *testing.T) {
	t.Parallel()
	if got := normalizeBotCommand("/start"); got != "/start" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeBotCommand("/start@urban_assist_bot"); got != "/start" {
		t.Fatalf("got %q", got)
	}
}

func TestCommandToAction(t *testing.T) {
	t.Parallel()
	action, ok := commandToAction("/start")
	if !ok || action != ActionHome {
		t.Fatalf("got %q ok=%v", action, ok)
	}
}

func TestIsCancelText(t *testing.T) {
	t.Parallel()
	if !isCancelText("отмена") || !isCancelText("/cancel") {
		t.Fatal("expected cancel")
	}
}
