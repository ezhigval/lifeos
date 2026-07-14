package telegram

import "testing"

func TestParseDeleteTarget(t *testing.T) {
	t.Parallel()
	if got := parseDeleteTarget("/delete @alice"); got != "alice" {
		t.Fatalf("got %q", got)
	}
	if got := parseDeleteTarget("/delete alice"); got != "alice" {
		t.Fatalf("got %q", got)
	}
	if got := parseDeleteTarget("/delete"); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := parseDeleteTarget("/delete me"); got != "me" {
		t.Fatalf("got %q", got)
	}
}

func TestIsDeleteConfirmText(t *testing.T) {
	t.Parallel()
	if !isDeleteConfirmText("confirm") || !isDeleteConfirmText("CONFIRM") || !isDeleteConfirmText("да") {
		t.Fatal("expected confirm texts")
	}
	if isDeleteConfirmText("maybe") {
		t.Fatal("unexpected")
	}
}

func TestCanDeleteUser(t *testing.T) {
	t.Parallel()
	h := &MessageHandler{adminTelegramID: 100}
	if !h.canDeleteUser(42, 42) {
		t.Fatal("self delete must be allowed")
	}
	if h.canDeleteUser(42, 99) {
		t.Fatal("non-admin must not delete others")
	}
	if !h.canDeleteUser(100, 99) {
		t.Fatal("admin must delete others")
	}
}
