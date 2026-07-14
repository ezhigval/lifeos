# TASK-001: Bugfix free-audit (zone only)

- **Agent:** frontend
- **Status:** OPEN
- **Priority:** P0
- **Stage:** bugfix
- **Owner ACK:** 2026-07-14 — no dogfood list; free audit; miniapp-ux merge in progress

## Goal

После merge miniapp-ux: найти и починить P0/P1 баги **только** в Mini App.

## In scope

- `web/miniapp/**`: auth/session/initData, routing, Home/Spheres/Finance/More/Habits/Calendar/Settings/Task detail
- Empty/error states, BottomNav, BackButton

## Out of scope

- Go backend / OpenAPI edits
- Telegram bot UI

## Acceptance

- [ ] P0 в зоне закрыты / BLOCKED с asks
- [ ] miniapp lint/build ok
- [ ] Report: `docs/agents/reports/frontend/TASK-001.md`
