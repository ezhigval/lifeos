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
	kb := telegram.MainReplyKeyboard("")
	if len(kb) != 4 {
		t.Fatalf("rows=%d want 4", len(kb))
	}
	if kb[3][0].Text != telegram.MenuSettings {
		t.Fatalf("last row=%v", kb[3])
	}
	kbApp := telegram.MainReplyKeyboard("https://example.com/app/")
	if len(kbApp) != 5 {
		t.Fatalf("rows=%d want 5 with mini app", len(kbApp))
	}
	last := kbApp[4][0]
	// Reply row must be plain text — web_app on reply KB yields empty initData on many clients.
	if last.Text != telegram.MenuMiniApp || last.WebApp != "" {
		t.Fatalf("mini app button=%+v", last)
	}
	open := telegram.InlineOpenMiniApp("https://example.com/app/")
	if len(open) != 1 || open[0][0].WebApp != "https://example.com/app/" {
		t.Fatalf("inline open=%+v", open)
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
