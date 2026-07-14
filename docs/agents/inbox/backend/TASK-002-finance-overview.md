# TASK-002: Finance overview API for Mini App

- **Agent:** backend
- **Status:** OPEN
- **Priority:** P1
- **Stage:** miniapp-functionality (Stage 2)
- **Source:** Frontend TASK-001 ask

## Goal

Add `GET /api/v1/finance/overview?period=YYYY-MM` so Mini App finance ring shows category breakdown (not cash-flow fallback with empty categories).

## Client contract (match exactly)

`period` query: `YYYY-MM` (e.g. `2026-07`), timezone of user for month bounds.

Response JSON:

```json
{
  "period_label": "июль 2026",
  "income_cents": 0,
  "expense_cents": 0,
  "net_cents": 0,
  "currency": "RUB",
  "categories": [
    { "name": "Еда", "amount_cents": 1000, "percent": 40.0, "color_hint": optional }
  ]
}
```

- `categories`: expense breakdown for that month (by category name); `percent` of total expenses (0–100).
- Empty month → zeros + `categories: []`.
- Invalid period → 400.

## In scope

- `internal/finance` app + infra + SQLC if needed
- HTTP route mount + runtime/api wiring
- OpenAPI + tests
- RU month label ok (`июль 2026` style) via user TZ

## Out of scope

- Mini App UI
- Telegram UX

## Acceptance

- [ ] Endpoint live and matches client type `FinanceOverview`
- [ ] OpenAPI synced
- [ ] Tests green
- [ ] Report: `docs/agents/reports/backend/TASK-002.md`
- [ ] Commit + push `cursor/docs-sync-orchestration-fe85`
