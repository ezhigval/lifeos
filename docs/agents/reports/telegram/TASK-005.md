# TASK-005 — Stage 3.0 Post-2.2 dogfood free-audit

**Agent:** telegram  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Base SHA:** `70f2641fe1e0123f2c673d19bb35aec57e2d40f3`  
**Scope:** `internal/transport/telegram/**`  
**Status:** DONE (P0/P1 fixed)

---

## Verdict

After ScheduleReminder → ReminderDTO: create path already compiled (`_, err :=`). Free-audit found **1 P0 + 2 P1** in cancel/FSM and reminder UX. Keyboard / Mini App entry still OK. Package tests green.

---

## Bugs found / fixed

| Sev | Area | Bug | Fix |
|-----|------|-----|-----|
| **P0** | Cancel / FSM (`/delete`) | `runAction` default used `SetStateOnly(Idle)` → kept `pending_delete_*` after reply-keyboard nav; later «да» could wipe the account. Non-confirm free text also fell through to intents while wipe stayed armed. | `SetState(Idle, basePayload)` drops wipe/draft keys; pending-delete branch requires confirm/cancel only |
| **P1** | Reminder create/cancel | Confirmations formatted `FireAt` in UTC (`15:04` / `15:04 02.01`) while calendar uses local TZ — «вечером» showed 16:00 for MSK | Use DTO + `formatLocalTime`; scheduled ack includes message |
| **P1** | Reminder create | `ParseFireAt` can yield a past wall-clock today («утром» after noon / «вечером» after 19:00) | `ensureFutureFireAt` rolls +24h until strictly future; empty message → RU hint |

---

## Verified OK (no fix needed)

| Area | Result |
|------|--------|
| ScheduleReminder DTO compile | `handler` already adapted in Stage 2.2 (`a1da86c`); now consumes `dto` for ack |
| Mini App entry | `MainReplyKeyboard` + URL rotate / `reply_kb_miniapp_url` / `ActionMiniApp` / `/start` clear unchanged |
| Reply keyboard | Sections + `web_app` shape; reinstall heuristic intact |
| Idle `/cancel` | «Нечего отменять.» short-circuit intact |
| Pending-delete cancel button | `cancelPendingDelete` still clears keys + home |

---

## Tests

```text
go test ./internal/transport/telegram/...   # OK
```

Coverage added: `ensureFutureFireAt`, `FormatReminderScheduled` message+at, basePayload drops pending-delete keys.

---

## Commits

- `a641bd2e66042e6790ea495532653e5822243554` — `fix(telegram): Stage 3.0 dogfood`
