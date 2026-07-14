# TASK-006 — Stage 3.1 Thin Telegram handler (strangler)

**Agent:** telegram  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Base SHA:** `6b300be63e19bf1e40eb75edd4ca3de30d89aeaf`  
**Scope:** `internal/transport/telegram/**`  
**Status:** DONE

---

## Verdict

Strangler extract of `runAction` and `dispatchIntent` (plus intent-only helpers) out of `handler.go`. No behaviour change; keyboard / Mini App / reminder wiring unchanged. Package tests green.

---

## Extraction

| File | Role |
|------|------|
| `actions.go` | `runAction` (menu/keyboard action switch + FSM side-effects) |
| `intent_dispatch.go` | `dispatchIntent` + resolve helpers (`resolveReminderToCancel`, note/contact/skill/sphere delete resolvers, project/sphere ID resolve, `ensureFutureFireAt`, `splitAndNames`, `atoiIntent`, `projectsBySphereView`) |
| `handler.go` | Update routing, present, views, callbacks, session/keyboard helpers |

No Backend facade needed.

---

## LOC

| File | Before | After |
|------|-------:|------:|
| `handler.go` | 1690 | 830 |
| `actions.go` | — | 82 |
| `intent_dispatch.go` | — | 812 |
| **handler.go Δ** | | **−860 (−51%)** |

---

## Tests

```text
go test ./internal/transport/telegram/...   # OK
go vet ./internal/transport/telegram/...    # OK
```

---

## Commits

- `b65f6872d64dd7a8c96cdddc660d6b391acf34f3` — `refactor(telegram): extract intent/action dispatch from handler`
