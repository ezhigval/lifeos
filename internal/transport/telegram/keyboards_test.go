package telegram_test

import (
	"testing"

	"github.com/valentinezhov/lifeos/internal/transport/telegram"
)

func TestTextToAction(t *testing.T) {
	t.Parallel()

	action, ok := telegram.TextToAction(telegram.MenuTasksToday)
	if !ok || action != telegram.ActionTasksToday {
		t.Fatalf("got %q ok=%v", action, ok)
	}
	action, ok = telegram.TextToAction(telegram.MenuSettings)
	if !ok || action != telegram.ActionSettings {
		t.Fatalf("settings got %q ok=%v", action, ok)
	}
}

func TestMainReplyKeyboardSections(t *testing.T) {
	t.Parallel()
	kb := telegram.MainReplyKeyboard()
	if len(kb) != 4 {
		t.Fatalf("rows=%d want 4", len(kb))
	}
	if kb[3][0] != telegram.MenuSettings {
		t.Fatalf("last row=%v", kb[3])
	}
}

func TestPrependInline(t *testing.T) {
	t.Parallel()
	actions := telegram.InlineHomeActions()
	got := telegram.PrependInline(actions, nil)
	if len(got) != len(actions) {
		t.Fatalf("got %d want %d", len(got), len(actions))
	}
}
