# TASK-009: Test coverage + observability debt (~20%)

- **Agent:** backend
- **Status:** DONE
- **Priority:** P1
- **Stage:** debt

## Goal

Подтянуть покрытие `domain`/`app` ближе к ≥80% на ключевых контекстах и укрепить observability defaults.

## In scope

1. **Tests** — raise coverage especially where low now:
   - `internal/finance/domain` (~18%)
   - `internal/finance/app` (~26%)
   - `internal/projects/domain` (~33%)
   - `internal/spheres/domain` (~42%)
   - `internal/tasks/app` (~39%)
   - keep `tasks/domain`, `identity/domain`, `ai/rulebased` healthy
2. **CI** — explicit cover gate for `domain`+`app` packages (script or grep thresholds); dedupe duplicate Integration test steps in `.github/workflows/ci.yml` if still duplicated
3. **Observability** — ensure `/metrics` documented; prometheus scrape config matches; optional smoke test that Init OTel disabled path works; short note in README or docs

## Out of scope

- Mini App React
- Telegram UX
- Requiring Grafana in default compose

## Acceptance

- [x] Measurable coverage lift on low packages
- [x] CI gate fails if critical packages drop too low
- [x] Report `docs/agents/reports/backend/TASK-009.md`
- [x] Commit + push
