# Milestones

**Version:** 0.4  
**See also:** [ROADMAP.md](ROADMAP.md) · [agents/CURRENT_STATE.md](../agents/CURRENT_STATE.md)

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
- [x] Tasks lifecycle: create / list / complete / cancel / edit / reschedule
- [x] Task fields: title, created_at, due_date, duration, projects, hashtags
- [x] Domain events
- [x] Integration tests (testcontainers)

---

## M3: Telegram Interface ✅

- [x] Long polling (+ webhook optional)
- [x] Rule-based intent resolver + golden tests
- [x] HTML formatting (RU)
- [x] E2E smoke tests
- [x] Reply keyboard + screen/dashboard model
- [x] FSM drafts (spheres/projects), `/clear`, `/delete`

---

## M4: Scheduler & Reviews ✅

- [x] Unified scheduled_jobs
- [x] Reminders, morning/evening reviews
- [x] Day availability + triage
- [x] Auto-reschedule incomplete tasks on evening review (+ notify)

---

## M5: Hardening 🚧

- [x] OpenTelemetry hooks
- [x] Observability compose profile
- [x] Backup scripts
- [ ] 14-day dogfooding gate — [DOGFOOD.md](DOGFOOD.md)

---

## M6: Finance ✅

- [x] Transactions, categories, debts
- [x] NL capture via intents
- [x] REST finance endpoints
- [x] Mini App finance card + sheet (client)

---

## M7: Habits + Calendar + Projects ✅

- [x] Habits + streaks
- [x] Calendar events
- [x] Projects + spheres (goals migrated 00022–00024)
- [x] task_projects M:N

---

## M8: Analytics ✅

- [x] Productivity summary query
- [x] Weekly/monthly reviews

---

## M9: API + Mini App (current)

| Item | Status |
|------|--------|
| REST `/api/v1` + JWT | ✅ |
| `POST /auth/telegram-webapp` | ✅ |
| OpenAPI sync | 🚧 improved; keep in parity with router |
| Ollama LLM resolver | 🚧 optional |
| Mini App | 🚧 merged UX: Home/Spheres/Finance/More/Habits/Calendar/Settings/Task detail — see [UX_UI_PLAN](../miniapp/UX_UI_PLAN.md) |

---

## M10+: Next

Stage 1 bugfix → Stage 2 Mini App + functionality (primary) → Intelligence later.
