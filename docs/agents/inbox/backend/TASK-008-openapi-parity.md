# TASK-008: OpenAPI ↔ router parity check

- **Agent:** backend
- **Status:** DONE
- **Priority:** P1
- **Stage:** debt

## Goal

Add a maintainable check that `/api/v1` routes in `router.go` Mount() stay in sync with `docs/api/openapi.yaml`, and wire it into Make/CI.

## In scope

- Go test under `internal/transport/http/api` (regex extract, no new OpenAPI validator deps)
- `make openapi-check`
- CI step after Format check
- Document any parity gaps in OpenAPI

## Out of scope

- Mini App / Telegram UX
- Full OpenAPI schema validation

## Acceptance

- [x] `make openapi-check` passes
- [x] CI runs the check
- [x] Report `docs/agents/reports/backend/TASK-008.md`
- [x] Commit + push
