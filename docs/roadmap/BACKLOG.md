# Backlog

**Version:** 0.3  
**See also:** [ROADMAP.md](ROADMAP.md)

Приоритет: **P0** (blocker) · **P1** (MVP) · **P2** (Phase 2) · **P3** (later)

---

## Epic: Infrastructure

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| INF-01 | Init Go module `github.com/valentinezhov/lifeos` | P0 | M1 |
| INF-02 | Docker Compose: **app + postgres** (default) | P0 | M1 |
| INF-03 | Goose migrations bootstrap | P0 | M1 |
| INF-04 | SQLC config + first queries | P0 | M1 |
| INF-05 | Cobra CLI: serve, migrate, version | P0 | M1 |
| INF-06 | slog JSON logging | P1 | M1 |
| INF-07 | Prometheus /metrics | P1 | M1 |
| INF-08 | OpenTelemetry tracing | P1 | M5 (Sprint 7) |
| INF-09 | GitHub Actions CI | P0 | M1 |
| INF-10 | golangci-lint + depguard | P1 | M1 |
| INF-11 | Air hot reload | P2 | M1 |
| INF-12 | Makefile (primary build contract) | P1 | M1 |
| INF-13 | Grafana dashboards | P2 | M5 (observability profile) |
| INF-14 | Testcontainers setup | P1 | M2 |
| INF-15 | Redis compose profile `cache` | P3 | Phase 2 |
| INF-16 | RUNBOOK + backup script | P1 | M5 |

## Epic: Identity

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| ID-01 | users table + migration | P0 | M2 |
| ID-02 | Telegram user allowlist | P0 | M3 |
| ID-03 | Seed default user | P0 | M2 |
| ID-04 | JWT generator/validator | P3 | Phase 3 |

## Epic: Settings

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| ST-01 | user_settings table | P1 | M2 |
| ST-02 | Default review times | P1 | M2 |
| ST-03 | Timezone-aware time parsing | P1 | M3 |

## Epic: Tasks

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| TK-01 | tasks table + migration (no project_id) | P0 | M2 |
| TK-02 | CreateTask use case | P0 | M2 |
| TK-03 | ListTasksToday use case | P0 | M2 |
| TK-04 | CompleteTask use case | P1 | M2 |
| TK-05 | RescheduleTasks use case | P1 | M3 |
| TK-06 | TriageOverloadedDay use case | P1 | M4 |
| TK-07 | Task domain unit tests | P0 | M2 |
| TK-08 | Task repo integration tests | P1 | M2 |

## Epic: Goals

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| GL-01 | goals table + migration | P1 | M2 |
| GL-02 | CreateGoal use case | P1 | M2 |
| GL-03 | GetGoalProgress use case | P1 | M3 |
| GL-04 | Link task to goal | P2 | M3 |

## Epic: Planning

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| PL-01 | day_availability table | P1 | M4 |
| PL-02 | SetDayAvailability use case | P1 | M4 |
| PL-03 | GetTopPriorities use case (query/) | P1 | M3 |

## Epic: Notifications

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| NT-01 | scheduled_jobs table (unified) | P0 | M4 |
| NT-02 | ScheduleReminder use case | P0 | M4 |
| NT-03 | Scheduler worker | P0 | M4 |
| NT-04 | MorningReview use case (query/) | P1 | M4 |
| NT-05 | EveningReview use case (query/) | P1 | M4 |
| NT-06 | TelegramNotifier adapter | P0 | M3 |
| NT-07 | Quiet hours check | P2 | M4 |

## Epic: AI

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| AI-01 | IntentResolver interface + typed intents | P0 | M3 |
| AI-02 | RuleBasedResolver | P0 | M3 |
| AI-03 | Time expression parser | P1 | M3 |
| AI-04 | Assistant template stub for reviews | P1 | M4 |
| AI-05 | Golden file tests for intents | P1 | M3 |
| AI-06 | LLM adapter | P3 | Phase 3 |

## Epic: Telegram Transport

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| TG-01 | Telegram Bot API client | P0 | M3 |
| TG-02 | Long polling loop | P0 | M3 |
| TG-03 | Update idempotency (processed_updates) | P0 | M3 |
| TG-04 | Message handler routing | P0 | M3 |
| TG-05 | HTML formatter | P1 | M3 |
| TG-06 | Inline keyboard callbacks | P2 | M4 |
| TG-07 | Webhook mode | P3 | Phase 1.5 |

## Epic: Events

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| EV-01 | domain_events table (+ source) | P1 | M2 |
| EV-02 | Event publisher in TX | P1 | M2 |

## Epic: Query (cross-context reads)

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| QR-01 | internal/query/ package scaffold | P1 | M3 |
| QR-02 | GetTopPriorities | P1 | M3 |
| QR-03 | MorningReview data assembly | P1 | M4 |
| QR-04 | EveningReview data assembly | P1 | M4 |

## Epic: Finance (Phase 2)

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| FN-01 | finance_transactions table | P2 | M6 |
| FN-02 | RecordIncome from NL | P2 | M6 |
| FN-03 | ListDebts query | P2 | M6 |
| FN-04 | CashFlowSummary | P2 | M6 |

## Epic: Habits (Phase 2)

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| HB-01 | habits + habit_logs tables | P2 | M7 |
| HB-02 | TrackHabit use case | P2 | M7 |

---

## Mini App (active)

План: [UX_UI_PLAN.md](../miniapp/UX_UI_PLAN.md)

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| MA-A1 | Auth `POST /auth/telegram-webapp` | P0 | A |
| MA-A2 | Telegram BackButton on nested routes | P0 | A |
| MA-A4 | Finance overview API + Mini App wire | P0 | A |
| MA-A3 | Tab «Ещё» + Settings stub | P1 | A |
| MA-B4 | Habits today + track | P1 | B |
| MA-B1 | CreateTask sheet (priority/due) | P1 | B |

Остальные MA-* — в UX_UI_PLAN §8.

---

## Icebox

- Multi-user SaaS
- Mobile app
- Google Calendar sync
- Bank API integration
- GraphQL API
- Voice messages → speech-to-text
