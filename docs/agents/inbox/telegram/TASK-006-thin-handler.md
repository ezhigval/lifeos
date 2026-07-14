# TASK-006: Stage 3.1 — Thin Telegram handler (strangler)

- **Agent:** telegram
- **Status:** DONE
- **Priority:** P1
- **Stage:** stabilization (3.1)

## Goal

Reduce `handler.go` god-object without big-bang rewrite. Extract intent dispatch and/or menu actions into separate files; keep behaviour identical.

## In scope

- `internal/transport/telegram/**`
- Extract `dispatchIntent` and/or `runAction` into `intent_dispatch.go` / `actions.go`
- Tests must stay green; no UX redesign

## Out of scope

- Moving business rules into transport
- Mini App
- Full rewrite of handler in one PR

## Acceptance

- [x] handler.go significantly smaller OR clear extracted modules
- [x] `go test ./internal/transport/telegram/...` green
- [x] Report `docs/agents/reports/telegram/TASK-006.md`
- [x] Commit + push
