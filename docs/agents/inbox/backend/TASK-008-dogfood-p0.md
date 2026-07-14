# TASK-008: Collect & fix dogfood P0 (backend)

- **Agent:** backend
- **Status:** DRAFT
- **Priority:** P0
- **Stage:** dogfood (G2→G3)

## Goal

Принять P0/P1 из 14-дневного dogfood Valentin ([DOGFOOD.md](../../../roadmap/DOGFOOD.md)) в зоне backend (API, auth, domains, scheduler/reviews) и закрыть блокеры gate.

## In scope

- Баги из `BUG-DOGFOOD-*` / asks с Agent=backend
- Регрессии API / timezone / finance / habits / reminders / reviews
- Auth `telegram-webapp` + JWT, если падает verify/stack

## Out of scope

- Mini App UI / Telegram keyboard presentation (чужие зоны)
- OpenAPI CI / observability budget (отдельный debt)

## Acceptance

- [ ] Architect → Status OPEN когда есть конкретный P0 backlog
- [ ] P0 в зоне fixed или BLOCKED с ask
- [ ] Report `docs/agents/reports/backend/TASK-008.md`
- [ ] Commit + push
