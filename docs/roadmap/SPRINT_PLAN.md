# Sprint Plan

**Version:** 0.3  
**See also:** [ROADMAP.md](ROADMAP.md)

> Sprints 0–6 выполнены. Актуальный статус — [MILESTONES.md](MILESTONES.md).

---

## Sprint 0: Approval & Repo Init (3 days)

**Goal:** SPEC v0.2 утверждён, git init, среда готова.

| Day | Tasks |
|-----|-------|
| D1 | Final review SPEC + ARCHITECTURE |
| D2 | `git init`, Go module, folder scaffold, `.env.example` |
| D3 | Docker Compose (app + postgres), Goose bootstrap, CI skeleton |

**Output:** Empty compiles, CI runs, DB migrates.

---

## Sprint 1: Platform & Identity (1 week)

**Goal:** Runnable HTTP server с observability.

### Tasks
1. `internal/platform/config` — env loading
2. `internal/platform/postgres` — pgx pool
3. `internal/platform/logging` — slog setup
4. `internal/platform/ids` — typed ID aliases
5. `cmd/lifeos` — Cobra: serve, migrate, version
6. `internal/transport/http` — Chi: /health, /ready, /metrics
7. Migration: users, user_settings
8. `internal/identity` — domain + app + seed user
9. `internal/settings` — domain + app
10. Prometheus metrics middleware
11. GitHub Actions full pipeline
12. Makefile: dev, test, lint, migrate

### Acceptance
- `make dev` → server on :8080
- `make migrate-up` → tables created
- CI green
- **No Redis in default compose**

---

## Sprint 2: Tasks Domain (1 week)

**Goal:** Tasks в app layer, тестируемо без Telegram.

### Tasks
1. Migration: tasks, domain_events (+ source column)
2. SQLC queries: insert, list by date, update status
3. `internal/tasks/domain` — Task entity, invariants
4. `internal/tasks/app` — CreateTask, ListToday, CompleteTask
5. `internal/tasks/infra` — SQLC repo
6. `internal/platform/events` — transactional publisher
7. Unit tests: domain + app
8. Integration tests: testcontainers
9. CLI debug: `lifeos task create "test"`

### Acceptance
- ≥ 80% coverage tasks domain + app
- Events written on create/complete

---

## Sprint 3: Goals + Planning Queries (1 week)

**Goal:** Goals и приоритизация.

### Tasks
1. Migration: goals
2. CreateGoal, GetGoalProgress use cases
3. `internal/query/` — GetTopPriorities
4. GoalReader port in tasks/app
5. Tests

### Acceptance
- Progress query: target, current, remaining
- Priorities ranked by urgency + due date

---

## Sprint 4: Telegram Transport (1 week)

**Goal:** Бот живой, основные команды работают.

### Tasks
1. `internal/transport/telegram` — Bot API client
2. Long polling + graceful shutdown
3. `processed_updates` idempotency (PG)
4. Allowlist middleware
5. `internal/ai` — IntentResolver + typed intents
6. RuleBasedResolver: 8 intents
7. Golden file tests (`testdata/`)
8. Time parser: «вечером», «завтра», «в 15:00»
9. HandleIncomingMessage orchestrator
10. HTML response formatter (RU)
11. Wire in serve.go

### Intents (minimum)
- task.create, task.list_today
- goal.create, goal.progress
- query.priorities
- reminder.create
- plan.set_availability
- unknown → fallback

### Acceptance
- E2E smoke: Telegram message → task in DB
- Response < 2s p95

---

## Sprint 5: Scheduler & Reminders (1 week)

**Goal:** Проактивные уведомления.

### Tasks
1. Migration: scheduled_jobs (unified, no reminders table)
2. `internal/platform/scheduler` — ticker worker
3. ScheduleReminder use case
4. TelegramNotifier port + adapter
5. Scheduler: process pending jobs FOR UPDATE SKIP LOCKED
6. Tests with fake clock

### Acceptance
- «Напомни через 2 минуты» → push через ~2 min

---

## Sprint 6: Reviews & Triage (1 week)

**Goal:** Daily rhythm + overload handling.

### Tasks
1. Migration: day_availability
2. `internal/query/` — MorningReview, EveningReview
3. Assistant template stub
4. Register review jobs in scheduler
5. SetDayAvailability use case
6. TriageOverloadedDay use case
7. Inline keyboard: confirm defer
8. Quiet hours check

### Acceptance
- Morning review at 08:00 (MSK)
- «Сегодня полный завал» → triage proposal
- «Работаю до 15:00» → saved

---

## Sprint 7: Hardening (1 week)

**Goal:** Production-ready для daily use.

### Tasks
1. Error handling audit (`%w` everywhere)
2. OpenTelemetry traces on HTTP + DB
3. Retry logic Telegram outbound
4. Backup script (pg_dump) + restore drill
5. docker restart policy doc
6. Coverage gap fill
7. Module READMEs for all MVP contexts
8. Grafana observability profile
9. Dogfooding bug fixes

### Acceptance
- 14 days dogfooding
- No P0 bugs

---

## Velocity Assumptions

- 1 developer, part-time ~15-20h/week
- Sprint = 1 calendar week
- MVP (Sprint 1-7) ≈ 7-8 weeks

---

## After MVP

Phase 2 backlog (Finance → M6). Re-plan based on actual velocity.
