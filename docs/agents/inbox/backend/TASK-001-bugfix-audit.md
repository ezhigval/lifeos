# TASK-001: Bugfix audit (zone only)

- **Agent:** backend
- **Status:** DRAFT
- **Priority:** P0
- **Stage:** bugfix

## Goal

Найти и починить P0/P1 баги **только** в backend-зоне после подтверждения списка владельцем / Architect.

## In scope

- `internal/<ctx>/{domain,app,infra}`, `query`, `ai`, `platform`, HTTP API, migrations/sqlc
- Регрессии auth/JWT/initData server-side, finance/tasks/projects use cases
- Синхронизация OpenAPI при фиксе контракта

## Out of scope

- `web/miniapp/**`
- Telegram buttons/FSM/keyboard cosmetics

## Acceptance

- [ ] Owner/Architect перевёл статус в `OPEN` и приложил конкретный bug list (или разрешил free audit)
- [ ] P0 в зоне закрыты или задокументированы как BLOCKED с причиной
- [ ] Tests for touched packages green
- [ ] Report: `docs/agents/reports/backend/TASK-001.md`
