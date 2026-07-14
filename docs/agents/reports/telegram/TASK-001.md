# TASK-001 — Telegram bugfix free-audit

**Agent:** telegram  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Scope:** `internal/transport/telegram/**`, `queries/telegram_sessions.sql`, `internal/notifications/infra/telegram_notifier.go`  
**Status:** DONE (P0/P1 in-zone closed)

---

## Verdict

Post–task-lifecycle merge: cancel/reschedule/tag intents and MiniAppURL wiring were present. Free audit found **1 P0 + 4 P1** transport/UX bugs (session dashboard clear, Mini App keyboard refresh, FSM idle cancel, project cancel buttons, notifier empty chat). Package tests green.

---

## Bugs found / fixed

| Sev | Area | Bug | Fix |
|-----|------|-----|-----|
| **P0** | Session / `/start` keyboard | `UpsertTelegramSession` used `COALESCE(EXCLUDED.dashboard_message_id, …)` while `Int8Nullable(0)` → NULL, so `resetReplyKeyboardFlag` could not clear the dashboard pointer | Write `dashboard_message_id = EXCLUDED.dashboard_message_id` (incl. NULL); `/start` payload clear always runs |
| **P1** | Mini App reply keyboard | After `LIFEOS_MINIAPP_URL` is set/cleared, existing sessions kept `reply_kb_set` + same version → edit-only path never attached/removed the `web_app` row | Track `reply_kb_miniapp`; bump `replyKBVersion` to 5; force reinstall on mismatch |
| **P1** | FSM `/cancel` | Idle `/cancel` / «отмена» fell through to intent resolver | Short-circuit: «Нечего отменять.» when idle (pending-delete cancel unchanged) |
| **P1** | Task lifecycle UX | Project task inline row only had ✓ and ignored cancelled; cancel callback always refreshed «today» | Reuse `taskDoneButtons` (✓+✕, skip done/cancelled); refresh project view when `view_project_id` set |
| **P1** | Notifier delivery | `chatID == 0` (lookup miss) still hit Bot API → noisy failures / retries | Skip send with warn when chat id empty |

### Hardened

- `FormatTaskCreated` shows duration/tags when domain extracted them (hashtags).
- `basePayload` preserves `reply_kb_miniapp`.

---

## Verified OK (no fix needed)

- **Reply keyboard attach:** Send dashboard + follow-up edit for inline (no ⌨️ carrier wipe).
- **`/clear` / `/delete`:** Reset SQL NULLs dashboard; wipe window + home re-present; admin/self delete confirm paths.
- **Lifecycle intents:** `task.cancel` / `task.reschedule_one` / `task.list_by_tag` → use cases; today list ✕ → `CancelTask`.
- **Mini App button markup:** `web_app` URL shape covered by keyboard tests; menu button set in `serve` when URL set.
- **Formatter tags/duration** on today list; runtime wires Cancel/Reschedule/ListByTag + MiniAppURL.

---

## Asks (out of scope — not blocked)

| Ask | Agent | Why |
|-----|-------|-----|
| Escape task titles in review `Summarize` before HTML notify | Backend / query | Template assistant concatenates raw titles into HTML evening/morning messages |
| Reminder plain-text vs HTML parse mode split | Backend / cmd | Notifier `Send` always uses client HTML path; reminder payloads are plain |

---

## Tests

```text
go test ./internal/transport/telegram/...   # OK
go build ./cmd/lifeos/...                   # OK
```

---

## Commits

_(filled after commit)_
