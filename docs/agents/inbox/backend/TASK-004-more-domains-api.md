# TASK-004: Stage 2.2 — API contracts for More domains

- **Agent:** backend
- **Status:** DONE
- **Priority:** P1
- **Stage:** miniapp-depth (Stage 2.2)

## Goal

Выровнять HTTP контракты Notes / Health / Career / Reminders / Debts / Analytics под Mini App `types.ts` + `client.ts`.

## In scope

- Handlers + app for those domains
- Status codes (404 missing entities, 400 validation)
- OpenAPI sync for touched routes
- Tests for changes

## Out of scope

- Mini App React
- Telegram UX

## Checklist vs client

- [x] notes list/create/delete
- [x] health weight/steps/sleep latest + record (404 on empty latest is OK if client handles)
- [x] career contacts/skills CRUD
- [x] reminders list/create/delete
- [x] finance debts list/create/pay
- [x] analytics/summary percent fields 0–100 ints

## Acceptance

- [x] Contracts aligned or mismatches documented for Frontend
- [x] Tests green
- [x] Report: `docs/agents/reports/backend/TASK-004.md`
- [x] Commit + push
