# TASK-004: Stage 2.2 — API contracts for More domains

- **Agent:** backend
- **Status:** OPEN
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

- notes list/create/delete
- health weight/steps/sleep latest + record (404 on empty latest is OK if client handles)
- career contacts/skills CRUD
- reminders list/create/delete
- finance debts list/create/pay
- analytics/summary percent fields 0–100 ints

## Acceptance

- [ ] Contracts aligned or mismatches documented for Frontend
- [ ] Tests green
- [ ] Report: `docs/agents/reports/backend/TASK-004.md`
- [ ] Commit + push
