# TASK-003 — Stage 2.1 Telegram Mini App entry sanity

**Agent:** telegram  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Scope:** `internal/transport/telegram/**` (menu button wiring in `cmd/lifeos/cmd/serve.go` read-only)  
**Status:** DONE — verified OK (no code change)

---

## Verdict

After URL-tracking (`reply_kb_miniapp_url`) + TASK-001/002 hardening: Mini App entry via `/start`, reply `web_app`, and chat menu button is stable. No bugs found. Package tests green.

---

## Audit checklist

| Path | Result |
|------|--------|
| `/start` → `resetReplyKeyboardFlag` | Clears `reply_kb_set`, `reply_kb_ver`, `reply_kb_miniapp`, `reply_kb_miniapp_url` + `DashboardMessageID=0` → next `Screen.Show` force-sends keyboard with current `h.miniAppURL` |
| `basePayload` | Preserves all `reply_kb_*` incl. `reply_kb_miniapp_url` (+ `view_project_id`); triage/`present` merge uses it — no wipe |
| URL rotate heuristic | `replyKeyboardInstalled` compares stored URL vs `MainReplyKeyboard` want URL; mismatch → force new dashboard + `web_app` reattach (no `/start` strictly required for reply KB) |
| Reply keyboard markup | `web_app` row only when `LIFEOS_MINIAPP_URL` non-empty; shape covered by `client_keyboard_test` / `keyboards_test` |
| Chat menu button | `SetChatMenuButton` on serve when URL set (`serve.go`) |
| ActionMiniApp text fallback | Points user to reply keyboard button; empty URL → clear RU error |

---

## Bugs found / fixed

None.

---

## Ops note (tunnel rotate)

1. Restart bot/`serve` so `LIFEOS_MINIAPP_URL` and **menu button** pick up the new hostname.
2. In Telegram: **`/start`** (or any action that resends the dashboard) so the reply `web_app` button gets the new URL. Auto-reattach also fires on the next `Screen.Show` when stored `reply_kb_miniapp_url` ≠ current env URL.
3. See also `docs/miniapp/LOCAL_DEV.md` (1033 / stale tunnel).

---

## Tests

```text
go test ./internal/transport/telegram/...   # OK
```

Existing coverage sufficient: `TestReplyKeyboardInstalled` (URL rotate), `TestReplyKeyboardHasMiniApp`, `TestReplyKeyboardMarkupShape`, `TestMainReplyKeyboardSections`.

---

## Commits

- _(none — verified OK, no code change)_
