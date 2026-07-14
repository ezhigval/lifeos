# TASK-001: Bugfix audit (zone only)

- **Agent:** frontend
- **Status:** DRAFT
- **Priority:** P0
- **Stage:** bugfix

## Goal

Найти и починить P0/P1 баги **только** в Mini App после подтверждения списка владельцем / Architect.

## In scope

- `web/miniapp/**`: auth/session/initData, routing, Home/Spheres/Finance UX, empty/error states

## Out of scope

- Go backend / OpenAPI edits
- Telegram bot UI

## Acceptance

- [ ] Status `OPEN` + bug list or free-audit ACK
- [ ] P0 в зоне закрыты / BLOCKED с cross-zone asks
- [ ] `npm` lint/build (или эквивалент Makefile) для miniapp
- [ ] Report: `docs/agents/reports/frontend/TASK-001.md`
