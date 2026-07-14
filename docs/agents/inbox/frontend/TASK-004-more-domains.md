# TASK-004: Stage 2.2 — Mini App More domains polish

- **Agent:** frontend
- **Status:** DONE
- **Priority:** P1
- **Stage:** miniapp-depth (Stage 2.2)

## Goal

Довести экраны «Ещё» (кроме уже закрытых Habits/Calendar/Settings) до ежедневной пригодности: Notes, Health, Career, Reminders, Debts, Analytics.

## In scope

- `web/miniapp/**` only
- Notes / Health / Career / Reminders / Debts / Analytics pages
- Create/delete flows, empty/error/loading, optimistic where safe, haptics
- Field mapping vs API; visible validation errors
- More hub grouping (daily vs rest) if still muddy
- Keep auth/`freezeInitData`/BrowserRouter intact

## Out of scope

- Go / OpenAPI / Telegram bot
- Full bot NL parity

## Acceptance

- [x] Listed screens usable without dead ends
- [x] `npm run build` (+ lint) green
- [x] Report: `docs/agents/reports/frontend/TASK-004.md`
- [x] Commit + push `cursor/docs-sync-orchestration-fe85`
