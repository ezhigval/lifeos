# TASK-009: Telegram package test hygiene

- **Agent:** telegram
- **Status:** DONE
- **Priority:** P2
- **Stage:** debt

## Goal

Добавить/усилить тесты на extracted `actions.go` / `intent_dispatch.go` узкие ветки (без UX redesign).

## In scope

- `internal/transport/telegram/*_test.go`
- Cover cancel/delete-nav and reminder formatting regressions from Stage 3.0 if gaps remain

## Out of scope

- Mini App, domain business rules

## Acceptance

- [x] `go test ./internal/transport/telegram/...` green with more assert coverage
- [x] Report `docs/agents/reports/telegram/TASK-009.md`
- [x] Commit if meaningful
