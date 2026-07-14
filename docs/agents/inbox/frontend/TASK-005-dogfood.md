# TASK-005: Stage 3.0 — Post-2.2 dogfood free-audit

- **Agent:** frontend
- **Status:** OPEN
- **Priority:** P0
- **Stage:** stabilization (3.0)

## Goal

Free-audit Mini App after Stage 2.2. Find and fix P0/P1 bugs in real daily flows: Home, Spheres, Finance, More domains, auth/session.

## In scope

- `web/miniapp/**`
- Auth/initData/session regressions
- Broken create/track/delete flows, bad field mapping, dead ends
- Keep BrowserRouter + freezeInitData

## Out of scope

- Go / Telegram bot

## Acceptance

- [ ] P0/P1 in zone fixed or BLOCKED with ask
- [ ] build/lint green
- [ ] Report `docs/agents/reports/frontend/TASK-005.md`
- [ ] Commit + push
