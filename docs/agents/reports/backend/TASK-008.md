# Report TASK-008

- **Agent:** backend
- **Status:** DONE
- **Branch / commits:** `cursor/docs-sync-orchestration-fe85` @ `1fa466c` (`1fa466cb9fa0edb6dde851c5fb069dce9dac9bf9`)

## Summary

Added OpenAPI ↔ router parity via `TestOpenAPIParity` (path + HTTP method). Wired as `make openapi-check` and a CI step after Format check. Current route lists already matched — no OpenAPI path docs needed.

## Changes

- `internal/transport/http/api/openapi_parity_test.go` — extract Mount() `r.Get|Post|…` routes and `paths:` `/api/v1/…` ops from YAML; fail on either direction mismatch
- `Makefile` — `openapi-check` target; included in `ci`
- `.github/workflows/ci.yml` — `make openapi-check` after Format check

## Parity result

| Check | Result |
|-------|--------|
| Paths router → OpenAPI | 50 / 50 |
| Paths OpenAPI → router | 50 / 50 |
| Method+path ops | match |

**Intentional mismatches:** none. Non-`/api/v1` mounts (health, metrics, static Mini App) are out of scope by design.

## Tests run

```
go test ./internal/transport/http/api -run TestOpenAPIParity -count=1
```

PASS.
