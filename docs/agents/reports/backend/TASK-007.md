# Report TASK-007

- **Agent:** backend
- **Status:** DONE
- **Branch / commits:** `cursor/docs-sync-orchestration-fe85` @ `e882ce1` (`e882ce186067ddacf7123d46cee08008b1b07fe8`)

## Summary

Optional LLM path is production-ready: **rule-based first**, Ollama only for unknown intents / review prose when `LIFEOS_LLM_ENABLED=true`. Down/timeouts degrade silently. Review summaries sanitize HTML (exact `<b>` only) and fall back to `templateassistant`.

## Strategy

| Surface | Order | Degrade |
|---------|-------|---------|
| IntentResolver | rulebased → (optional) Ollama on `unknown` | Ollama error/timeout/unknown → primary unknown |
| Assistant | (optional) Ollama → templateassistant | error/empty/timeout/unsafe → template |

HTTP timeout for Ollama: **8s** (was 30s). Composite also applies its own deadline (8s resolve / 15s summarize).

## Changes

- `internal/ai/composite/{resolver,assistant}.go` — timeouts, docs, safe summarize
- `internal/ai/reviewsafe/` — Telegram HTML sanitize (`<b>` only, attrs rejected, length cap)
- `internal/ai/ollama/client.go` — 8s default; `NewClientWithTimeout`
- `internal/ai/ollama/assistant.go` — sanitize before return
- `internal/ai/templateassistant/` — `EscapePlain` on titles/projects
- `cmd/lifeos/cmd/{resolver,assistant}.go` — wiring comments (behavior unchanged)
- `.env.example`, `docs/architecture/ARCHITECTURE.md`, `docs/adr/005-…` — LIFEOS_LLM note
- Failure-path tests: composite + ollama httptest down/timeout + HTML safety

## Tests run

```
go test ./internal/ai/...
go build ./cmd/lifeos
```

All green.

## Cross-zone asks

None. Telegram TASK-007 may verify review HTML still parses; Backend already escapes/sanitizes before send.
