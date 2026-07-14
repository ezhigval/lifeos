# TASK-009 — Telegram package test hygiene

**Agent:** telegram  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Base SHA:** `f93d3e9a04a5061ac4288bc50aeccdc353965572`  
**Scope:** `internal/transport/telegram/**`  
**Status:** DONE

---

## Verdict

Hardened Stage 3.0 regression coverage for delete-nav (`filterBasePayload`) and reminder TZ/format paths (`formatLocalTime`, cancel ack, `matchReminderToCancel`, `ensureFutureFireAt` edges). Pure helpers extracted so tests stay fast (no DB). Package coverage **10% → 13%**.

---

## Gaps closed

| Area | Before | After |
|------|--------|-------|
| Delete-nav / `basePayload` | Mimicked filter in test; production path untested | `filterBasePayload` extracted + tested (drops `pending_delete_*` / drafts, keeps KB/view keys) |
| Reminder TZ (`formatLocalTime`) | 0% — Stage 3.0 MSK bug unguarded | MSK 16:00 UTC → `14.07 19:00`; invalid TZ → UTC |
| Reminder cancel format | 0% | `FormatReminderCancelled` / `FormatReminderNotFound` |
| Reminder match | 0% (`resolveReminderToCancel` DB-bound) | `matchReminderToCancel` pure helper + empty/hint/miss/bad-uuid |
| `ensureFutureFireAt` | Past→tomorrow only | Also equal-now roll + multi-day stale |
| Intent helpers | 0% | `splitAndNames`, `atoiIntent` |
| Calendar create ack TZ | 0% | `FormatCalendarEventCreated` MSK local |

---

## Tests

```text
go test ./internal/transport/telegram/... -cover   # OK, ~13%
go vet ./internal/transport/telegram/...           # OK
```

New/updated: `intent_dispatch_test.go`, `formatter_test.go`, `delete_user_test.go`.

---

## Commits

- *(this commit)* — `test(telegram): harden Stage 3 regression coverage`
