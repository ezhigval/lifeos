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
	if _, ok := commandToAction("/clear"); ok {
		t.Fatal("/clear must not map via commandToAction; handled separately")
	}
	if got := normalizeBotCommand("/clear@lifeos_bot"); got != "/clear" {
		t.Fatalf("got %q", got)
	}
}

func TestIsCancelText(t *testing.T) {
	t.Parallel()
	if !isCancelText("отмена") || !isCancelText("/cancel") {
		t.Fatal("expected cancel")
	}
}
