# TASK-005: Stage 3.0 — Post-2.2 dogfood free-audit

- **Agent:** telegram
- **Status:** OPEN
- **Priority:** P0
- **Stage:** stabilization (3.0)

## Goal

Free-audit bot UX after reminder Execute signature change and Stage 2. Fix P0/P1: keyboard, Mini App entry, reminder/create intents, cancel/FSM.

## In scope

- `internal/transport/telegram/**`, TG notifier
- Compile/runtime after ScheduleReminder returns DTO

## Out of scope

- Mini App React, AI prompts

## Acceptance

- [ ] tests green
- [ ] Report `docs/agents/reports/telegram/TASK-005.md`
- [ ] Commit + push if fixes
