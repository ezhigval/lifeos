# TASK-004: Stage 2.2 — Telegram light check

- **Agent:** telegram
- **Status:** OPEN
- **Priority:** P3
- **Stage:** miniapp-depth (Stage 2.2)

## Goal

Не блокировать Stage 2.2. Быстрый sanity: Mini App entry still OK; formatter/intents for notes/health/career/reminders/debts still wire to existing use cases if touched recently. No big refactors.

## In scope

- `internal/transport/telegram/**` only if a clear P1 bug
- Otherwise report verified OK

## Out of scope

- Mini App React
- Thin-handler rewrite

## Acceptance

- [ ] Tests green if code touched
- [ ] Report: `docs/agents/reports/telegram/TASK-004.md`
- [ ] Commit only if fixes
