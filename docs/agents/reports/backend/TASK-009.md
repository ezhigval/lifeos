# Report TASK-009

- **Agent:** backend
- **Status:** DONE
- **Branch / commits:** `cursor/docs-sync-orchestration-fe85` @ `bbcd467` (`bbcd467daffd99c53ff93dff858df9c65e9d5a74`)

## Summary

Raised unit coverage on previously weak `domain`/`app` packages, added `scripts/check-coverage.sh` + `make coverage-check` CI gate, removed the duplicate Integration test step in GitHub Actions, and documented `/metrics` + `make observability-up`.

## Coverage before → after

| Package | Before | After | Gate min |
|---------|--------|-------|----------|
| `internal/finance/domain` | 18.2% | 97.7% | 80% |
| `internal/finance/app` | 26.3% | 78.4% | 70% |
| `internal/projects/domain` | 33.3% | 100% | 80% |
| `internal/spheres/domain` | 41.7% | 100% | 80% |
| `internal/tasks/app` | 39.4% | 72.8% | 65% |
| `internal/tasks/domain` | 78.5% | 78.5% | 75% |
| `internal/identity/domain` | 85.7% | 85.7% | 80% |
| `internal/ai/rulebased` | 69.6% | 69.6% | 65% |

## Changes

- Tests: finance domain (money/category/tx/debt), finance app (income/expense/cashflow/debts/pay-by-id), projects Progress, spheres Rename, tasks mutate/list/auto-reschedule/edit
- `internal/transport/http/metrics_route_test.go` — `/metrics` + `/health` register
- `scripts/check-coverage.sh` — parse coverprofile statement blocks vs minima
- `Makefile` — `coverage-check`; included in `ci`
- `.github/workflows/ci.yml` — Coverage gate after Test; single Integration test step (after migrate)
- README + `docs/architecture/ARCHITECTURE.md` — `/metrics`, prometheus scrape, `make observability-up`

## Observability

- `/metrics` already mounted via `promhttp` in `internal/transport/http/server.go`
- Scrape config `deployments/prometheus/prometheus.yml` → `app:8080` (matches compose)
- OTel disabled-path smoke already covered by `internal/platform/otel/tracing_test.go`

## Tests run

```
go test ./internal/finance/... ./internal/projects/domain ./internal/spheres/domain \
  ./internal/tasks/... ./internal/identity/domain ./internal/ai/rulebased \
  ./internal/transport/http -coverprofile=coverage.out
./scripts/check-coverage.sh coverage.out
```

PASS; coverage-check OK.
