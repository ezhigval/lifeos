# TASK-002: Align Telegram UX with Stage 2

- **Agent:** telegram
- **Status:** OPEN
- **Priority:** P2
- **Stage:** miniapp-functionality (Stage 2)

## Goal

Keep bot UX coherent after task lifecycle + Mini App entry. Fix residual keyboard/formatter/session gaps; ensure finance/task lifecycle intents still present cleanly.

## In scope

- `internal/transport/telegram/**`, TG notifier
- Reply keyboard / Mini App button stability
- Formatter for duration/tags/cancel/reschedule if gaps remain
- No domain rule invention — call existing use cases

## Out of scope

- Mini App React
- New finance overview business logic (Backend owns HTTP overview)

## Acceptance

- [ ] No keyboard/session P0 regressions
- [ ] `go test ./internal/transport/telegram/...` green
- [ ] Report: `docs/agents/reports/telegram/TASK-002.md`
- [ ] Commit + push branch
