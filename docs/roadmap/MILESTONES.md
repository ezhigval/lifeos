# Milestones

**Version:** 0.3  
**See also:** [ROADMAP.md](ROADMAP.md)

---

## M0: Architecture & Design ✅

**Status:** Complete

---

## M1: Project Skeleton ✅

- [x] Go module `github.com/valentinezhov/lifeos`
- [x] Folder structure per ARCHITECTURE.md
- [x] Docker Compose: app + postgres
- [x] Goose + SQLC
- [x] HTTP :8080 — health, ready, metrics
- [x] GitHub Actions CI
- [x] Makefile

---

## M2: Core Domain ✅

- [x] Users + settings
- [x] Tasks CRUD
- [x] Domain events
- [x] Integration tests (testcontainers)

---

## M3: Telegram Interface ✅

- [x] Long polling (+ webhook optional)
- [x] Rule-based intent resolver + golden tests
- [x] HTML formatting (RU)
- [x] E2E smoke tests

---

## M4: Scheduler & Reviews ✅

- [x] Unified scheduled_jobs
- [x] Reminders, morning/evening reviews
- [x] Day availability + triage

---

## M5: Hardening 🚧

- [x] OpenTelemetry hooks
- [x] Observability compose profile
- [x] Backup scripts
- [ ] 14-day dogfooding gate

---

## M6: Finance ✅

- [x] Transactions, categories, debts
- [x] NL capture via intents

---

## M7: Habits + Calendar + Projects ✅

- [x] Habits + streaks
- [x] Calendar events
- [x] Projects + spheres (goals migrated)
- [x] task_projects M:N

---

## M8: Analytics ✅

- [x] Productivity summary query
- [x] Weekly/monthly reviews

---

## M9–M10: Intelligence & API

| Item | Status |
|------|--------|
| REST `/api/v1` + JWT | ✅ |
| OpenAPI spec | ✅ (partial sync) |
| Ollama LLM resolver | 🚧 optional |
| Mini App UI | 🚧 scaffold → [UX/UI plan](../miniapp/UX_UI_PLAN.md) |

See [ROADMAP.md](ROADMAP.md) for Phase 3–4.
