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

func TestFilterBasePayloadDropsPendingDeleteKeys(t *testing.T) {
	t.Parallel()
	// runAction default path uses SetState(Idle, basePayload→filterBasePayload)
	// so walking away via reply keyboard must not keep pending_delete_* armed.
	payload := map[string]any{
		replyKBSetKey:            true,
		replyKBVersionKey:        float64(replyKBVersion),
		replyKBMiniAppKey:        false,
		replyKBMiniAppURLKey:     "https://example.test/app",
		"view_project_id":        "proj-1",
		PayloadPendingDeleteTG:   float64(42),
		PayloadPendingDeleteName: "@me",
		"draft_task_title":       "stale",
		"draft_project_name":     "wipe-me",
	}
	out := filterBasePayload(payload)
	if _, ok := out[PayloadPendingDeleteTG]; ok {
		t.Fatal("pending delete must not survive base payload rebuild")
	}
	if _, ok := out[PayloadPendingDeleteName]; ok {
		t.Fatal("pending delete name must not survive")
	}
	if _, ok := out["draft_task_title"]; ok {
		t.Fatal("draft keys must not survive base payload rebuild")
	}
	if _, ok := out["draft_project_name"]; ok {
		t.Fatal("draft project must not survive")
	}
	if out[replyKBSetKey] != true {
		t.Fatal("reply keyboard flags must be preserved")
	}
	if out[replyKBMiniAppURLKey] != "https://example.test/app" {
		t.Fatal("miniapp url must be preserved")
	}
	if out["view_project_id"] != "proj-1" {
		t.Fatal("view_project_id must be preserved")
	}
	if len(filterBasePayload(nil)) != 0 {
		t.Fatal("nil payload → empty map")
	}
}
