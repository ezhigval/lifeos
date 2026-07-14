# TASK-001: Bugfix free-audit (zone only)

- **Agent:** telegram
- **Status:** OPEN
- **Priority:** P0
- **Stage:** bugfix
- **Owner ACK:** 2026-07-14 — no dogfood list; free audit; task-lifecycle merge touch TG handler

## Goal

После merge task-lifecycle: найти и починить P0/P1 баги **только** в Telegram transport/UX.

## In scope

- `internal/transport/telegram/**`, TG notifier
- Keyboard/screen/FSM/commands; formatter + intents wiring for new task actions
- Не ломать reply keyboard / dashboard model

## Out of scope

- Domain rules inventing
- Mini App React
- AI resolver internals (кроме вызова port)

## Acceptance

- [ ] P0 в зоне закрыты / BLOCKED с asks к Backend
- [ ] Package tests green
- [ ] Report: `docs/agents/reports/telegram/TASK-001.md`
