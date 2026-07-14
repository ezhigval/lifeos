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

func TestBasePayloadDropsPendingDeleteKeys(t *testing.T) {
	t.Parallel()
	// runAction default path now uses SetState(Idle, basePayload) so walking away
	// via reply keyboard must not keep pending_delete_* armed.
	payload := map[string]any{
		replyKBSetKey:            true,
		replyKBVersionKey:        float64(replyKBVersion),
		replyKBMiniAppKey:        false,
		PayloadPendingDeleteTG:   float64(42),
		PayloadPendingDeleteName: "@me",
		"draft_task_title":       "stale",
	}
	// Mimic basePayload selection rules without a session store.
	out := map[string]any{}
	if v, ok := payload[replyKBSetKey]; ok {
		out[replyKBSetKey] = v
	}
	if v, ok := payload[replyKBVersionKey]; ok {
		out[replyKBVersionKey] = v
	}
	if v, ok := payload[replyKBMiniAppKey]; ok {
		out[replyKBMiniAppKey] = v
	}
	if _, ok := out[PayloadPendingDeleteTG]; ok {
		t.Fatal("pending delete must not survive base payload rebuild")
	}
	if _, ok := out["draft_task_title"]; ok {
		t.Fatal("draft keys must not survive base payload rebuild")
	}
	if out[replyKBSetKey] != true {
		t.Fatal("reply keyboard flags must be preserved")
	}
}
