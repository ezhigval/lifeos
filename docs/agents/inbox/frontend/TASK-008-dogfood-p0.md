# TASK-008: Collect & fix dogfood P0 (frontend)

- **Agent:** frontend
- **Status:** DRAFT
- **Priority:** P0
- **Stage:** dogfood (G2→G3)

## Goal

Принять P0/P1 из 14-дневного dogfood Valentin ([DOGFOOD.md](../../../roadmap/DOGFOOD.md)) в зоне Mini App и закрыть блокеры gate.

## In scope

- Баги из `BUG-DOGFOOD-*` / asks с Agent=frontend
- Home, Spheres, Finance ring, Habits, Calendar, Settings, Notes/Health/Career/Debts/Analytics/Reminders
- Auth/session / freezeInitData регрессии на клиенте

## Out of scope

- Go API / Telegram bot handlers
- Полный bot↔Mini App parity redesign

## Acceptance

- [ ] Architect → Status OPEN когда есть конкретный P0 backlog
- [ ] P0 в зоне fixed или BLOCKED с ask
- [ ] Report `docs/agents/reports/frontend/TASK-008.md`
- [ ] Commit + push
