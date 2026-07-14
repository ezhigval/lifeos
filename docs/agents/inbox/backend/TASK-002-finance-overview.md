# TASK-002: Finance overview API for Mini App

- **Agent:** backend
- **Status:** DRAFT
- **Priority:** P1
- **Stage:** miniapp-functionality (Stage 2)
- **Source:** Frontend TASK-001 ask

## Goal

Add `GET /api/v1/finance/overview?period=` (or equivalent) so Mini App finance ring can show category breakdown without cash-flow fallback.

## In scope

- finance app + HTTP + OpenAPI + tests
- Period semantics aligned with miniapp client expectations (`docs/miniapp` / client `FinanceOverview` type)

## Out of scope

- Mini App UI (Frontend wires once API exists)

## Acceptance

- [ ] Endpoint returns overview DTO Mini App can consume
- [ ] OpenAPI synced
- [ ] Tests green
- [ ] Report under reports/backend/
