# TASK-007: Stage 3.2 — Intelligence polish

- **Agent:** backend
- **Status:** OPEN
- **Priority:** P1
- **Stage:** intelligence (3.2)

## Goal

Make LLM path production-ready as optional: composite resolver fallback, assistant summaries quality/safety, tests, docs (LIFEOS_LLM_ENABLED).

## In scope

- `internal/ai/**`, wiring in `cmd/lifeos/cmd/{resolver,assistant}.go`
- Golden/unit tests; clear degrade to rulebased when Ollama down
- OpenAPI/docs note if needed

## Out of scope

- Requiring Ollama in default compose
- Mini App / Telegram UX redesign

## Acceptance

- [ ] Composite reliably falls back
- [ ] Tests cover failure paths
- [ ] Report `docs/agents/reports/backend/TASK-007.md`
- [ ] Commit + push
