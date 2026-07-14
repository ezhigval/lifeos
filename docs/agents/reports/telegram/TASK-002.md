# TASK-002 — Stage 2 Telegram UX align

**Agent:** telegram  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Scope:** `internal/transport/telegram/**` (notifier delivery unchanged)  
**Status:** DONE (no P0 left; small hardening shipped)

---

## Verdict

Post TASK-001 + Mini App / lifecycle merges: free-audit found **no P0**. Shipped **2 P2** transport hardenings (triage payload vs Mini App keyboard; formatter completeness) and **1 P2** stale ✕/✓ callback UX. Finance intents and notifier empty-chat skip verified OK. Package tests green.

---

## Bugs found / fixed

| Sev | Area | Bug | Fix |
|-----|------|-----|-----|
| **P2** | Reply keyboard / Mini App | `present()` stored triage `defer_tasks` as a fresh payload, wiping `reply_kb_*` → forced dashboard resend / `web_app` reinstall on every triage with lows | Merge `defer_tasks` into `basePayload` (keeps Mini App keyboard flags + `view_project_id`) |
| **P2** | Formatters | Tag/project lists omitted duration/tags; reschedule with nil due showed empty `→` | Shared `formatTaskMeta`; `FormatTasksByTag` / `FormatProjectTasks` / `FormatTaskRescheduled` completed |
| **P2** | Lifecycle callbacks | Stale ✓/✕ after done/cancelled showed raw wrapped errors | Map `ErrTaskNotFound` / `ErrAlreadyCancelled` / `ErrCannotCancelDone` / `ErrCannotComplete` → RU copy + list refresh |

---

## Verified OK (no code change)

- **Reply keyboard + `web_app`:** markup shape, `reply_kb_miniapp` reinstall heuristic, menu button wiring via `LIFEOS_MINIAPP_URL`
- **FSM `/cancel` / idle / pending-delete:** idle short-circuit; draft cancel → idle + «Отменено.»; delete confirm cancel path intact
- **`/clear` / `/start`:** session reset + keyboard force-reattach still correct after TASK-001 SQL upsert
- **Finance intents:** income/expense/debts/cash-flow/pay → existing use cases + formatters only (no overview HTTP)
- **Notifier:** `chatID == 0` skip still in place; no further delivery gaps in zone

### Residual / deferred

| Item | Sev | Notes |
|------|-----|-------|
| Thin `handler.go` strangler | P2 | Backlog TG-08; not blocking Stage 2 |
| Reminder HTML vs plain / review title escape | P2 | Out of zone (Backend / cmd asks from TASK-001) |

---

## Tests

```text
go test ./internal/transport/telegram/...   # OK
```

New coverage: `formatter_test.go`, `lifecycle_callback_test.go`.

---

## Commits

- _(this push)_ — `fix(telegram): Stage 2 UX align`
