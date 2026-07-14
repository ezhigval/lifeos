# TASK-001: Bugfix audit (zone only)

- **Agent:** telegram
- **Status:** DRAFT
- **Priority:** P0
- **Stage:** bugfix

## Goal

Найти и починить P0/P1 баги **только** в Telegram transport/UX после подтверждения списка владельцем / Architect.

## In scope

- `internal/transport/telegram/**`, TG notifier delivery
- Keyboard persistence, screen edit/send, FSM cancel, commands `/clear` `/delete`, callbacks

## Out of scope

- Domain/business rules inventing
- Mini App React
- AI resolver internals

## Acceptance

- [ ] Status `OPEN` + bug list or free-audit ACK
- [ ] P0 в зоне закрыты / BLOCKED с asks к Backend
- [ ] Package tests green
- [ ] Report: `docs/agents/reports/telegram/TASK-001.md`
