# TASK-001: Bugfix free-audit (zone only)

- **Agent:** backend
- **Status:** OPEN
- **Priority:** P0
- **Stage:** bugfix
- **Owner ACK:** 2026-07-14 — no dogfood list; free audit; WIP merges in progress on orchestrator branch

## Goal

После merge task-lifecycle (+ зависимости): найти и починить P0/P1 баги **только** в backend-зоне.

## In scope

- `internal/<ctx>/{domain,app,infra}`, `query`, `ai`, `platform`, HTTP API, migrations/sqlc
- Новый task lifecycle (duration/tags/edit/cancel/reschedule) — correctness + tests
- OpenAPI sync при изменении контракта
- API gaps нужные Mini App (если блокируют — минимальный фикс или ask в отчёте)

## Out of scope

- `web/miniapp/**`
- Telegram buttons/FSM/keyboard cosmetics

## Acceptance

- [ ] P0 в зоне закрыты или BLOCKED с причиной + cross-zone ask
- [ ] Touched package tests green
- [ ] Report: `docs/agents/reports/backend/TASK-001.md`
