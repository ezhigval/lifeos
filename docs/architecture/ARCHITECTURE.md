# LifeOS — Architecture (единый лист)

**Version:** 0.4  
**Status:** Current (synced to code, 2026-07-14)  
**Detail:** [DOMAIN_MODEL.md](DOMAIN_MODEL.md)

---

## 1. Что это

| | |
|---|---|
| **Стиль** | Modular Monolith |
| **Deploy** | Один binary (`lifeos`), один процесс |
| **Пользователи v1** | 1 (single-tenant, `user_id` на всех таблицах) |
| **UI v1** | Telegram (long polling) + REST `/api/v1` + Telegram Mini App |
| **Хранилище** | PostgreSQL 16 (source of truth) |
| **Цель** | Transport-agnostic бизнес-логика, готовая к REST/Web позже |

---

## 2. Почему Modular Monolith

| За | Против (принято) |
|----|------------------|
| Простой деплой на домашнем Mac | Horizontal scale ограничен (достаточно для 1 user) |
| Одна БД, низкая latency | Дисциплина import rules обязательна |
| Чёткие границы → выделение сервисов позже | |
| Нет operational overhead microservices | |

Microservices и «голый» monolith без границ — отклонены ([ADR-001](../adr/001-modular-monolith.md)).

---

## 3. Слои

```
┌─────────────────────────────────────────────────────────┐
│  TRANSPORT (thin adapters)                              │
│  internal/transport/telegram   internal/transport/http  │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│  APPLICATION (use cases)                                │
│  internal/<context>/app/                                  │
│  internal/query/          ← cross-context reads         │
│  internal/ai/             ← IntentResolver, Assistant   │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│  DOMAIN (entities, rules, domain errors)              │
│  internal/<context>/domain/                           │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│  INFRASTRUCTURE                                         │
│  internal/<context>/infra/   internal/platform/         │
└─────────────────────────────────────────────────────────┘
```

### Правила зависимостей

1. `domain` не импортирует `app`, `infra`, `transport`
2. `domain` imports: `time`, `errors`, `fmt`, `internal/platform/ids`
3. `app` импортирует `domain`; определяет reader ports у потребителя
4. `infra` реализует repo interfaces
5. `transport` вызывает `app` use cases, не содержит бизнес-логики
6. Cross-context: только UUID, без import чужого `domain`

---

## 4. Bounded Contexts

```
┌─────────────┐     ┌─────────────┐
│  identity   │────▶│  settings   │
└─────────────┘     └─────────────┘
       │
       ▼
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│    tasks    │◀───▶│  projects   │◀───▶│   spheres    │
└─────────────┘     └─────────────┘     └──────────────┘
       │                   │
       ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│  planning   │     │   finance   │     │    habits    │
└─────────────┘     └─────────────┘     └──────────────┘
       │                   │                    │
       └───────────────────┼────────────────────┘
                           ▼
                  ┌─────────────────┐
                  │  notifications  │  ← scheduled_jobs
                  └─────────────────┘
                           │
       ┌───────────────────┼───────────────────┐
       ▼                   ▼                   ▼
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│  calendar   │     │  knowledge  │     │    health    │
└─────────────┘     └─────────────┘     └──────────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │     career      │
                  └─────────────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │  ai (ports)     │  ← IntentResolver, Assistant
                  └─────────────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │  transport      │  ← telegram, http
                  └─────────────────┘

Cross-reads: internal/query/ (priorities, reviews, analytics)
```

| Context | Package | Aggregates / Entities |
|---------|---------|----------------------|
| identity | `internal/identity/` | User |
| settings | `internal/settings/` | UserSettings |
| tasks | `internal/tasks/` | Task (M:N projects via `task_projects`) |
| projects | `internal/projects/` | Project (M:N spheres) |
| spheres | `internal/spheres/` | LifeSphere |
| planning | `internal/planning/` | DayAvailability |
| finance | `internal/finance/` | Transaction, Debt, Category |
| habits | `internal/habits/` | Habit, HabitLog |
| calendar | `internal/calendar/` | CalendarEvent |
| knowledge | `internal/knowledge/` | Note |
| health | `internal/health/` | Weight, Steps, Sleep |
| career | `internal/career/` | Contact, Skill |
| notifications | `internal/notifications/` | ScheduledJob |
| ai | `internal/ai/` | Ports only |
| query | `internal/query/` | Read services |

> **Goals** удалены: иерархия целей перенесена в `projects` (миграция 00022–00024).

---

## 5. Структура репозитория

```
lifeos/
├── cmd/lifeos/
│   ├── main.go
│   └── cmd/                        # Cobra: serve, migrate, version, …
│       ├── serve.go                # composition root entry
│       ├── runtime.go              # manual DI wiring
│       └── api.go                  # HTTP API deps assembly
├── internal/
│   ├── platform/                   # config, postgres, auth, scheduler, events, ids, db(sqlc)
│   ├── identity/   {domain, app, infra}
│   ├── settings/   {domain, app, infra}
│   ├── tasks/      {domain, app, infra}
│   ├── projects/   {domain, app, infra}
│   ├── spheres/    {domain, app, infra}
│   ├── planning/   {domain, app, infra}
│   ├── finance/    {domain, app, infra}
│   ├── habits/     {domain, app, infra}
│   ├── calendar/   {domain, app, infra}
│   ├── knowledge/  {domain, app, infra}
│   ├── health/     {domain, app, infra}
│   ├── career/     {domain, app, infra}
│   ├── notifications/ {domain, app, infra}
│   ├── ai/         {rulebased, ollama, composite, templateassistant}
│   ├── query/                      # priorities, reviews, analytics
│   └── transport/
│       ├── http/                   # Chi server + /api/v1 (JWT)
│       └── telegram/               # polling/webhook, screens, keyboards, FSM
├── web/miniapp/                    # Telegram Mini App (React 19 + Vite)
├── migrations/                     # Goose SQL (00001–00026+)
├── queries/                        # SQLC sources
├── deployments/                    # Dockerfile, docker-compose, prometheus/grafana
├── e2e/
├── docs/                           # architecture, ADR, roadmap, agents, api
├── Makefile
├── sqlc.yaml
└── .github/workflows/ci.yml
```

---

## 6. Composition Root

`cmd/lifeos/cmd/runtime.go` (+ `serve.go`, `api.go`) — composition root ([ADR-008](../adr/008-manual-di.md)).

```go
// Псевдокод (актуально под projects/spheres)
db := postgres.New(cfg)
taskRepo := tasksinfra.NewRepository(db)
projectRepo := projectsinfra.NewRepository(db)
sphereRepo := spheresinfra.NewRepository(db)
intentResolver := ai.NewRuleBasedResolver() // + optional Ollama composite
eventPub := events.NewPublisher(db)

createTask := tasksapp.NewCreateTask(taskRepo, eventPub)
handleMessage := telegram.NewHandler(intentResolver, useCases, screen)

scheduler := platform.NewScheduler(db, clock)
scheduler.Register("morning_review", query.MorningReview)
scheduler.Register("reminder", notifapp.DeliverReminder)

httpSrv := httptransport.New(cfg, api.NewRouter(deps), staticMiniApp)
tgPoller := telegram.NewPoller(cfg, handleMessage)

// errgroup: http + poller + scheduler, graceful shutdown on SIGTERM
```

Revisit Wire/Fx only if composition root becomes unmanageable.

---

## 7. Request Flow — Telegram

```
Telegram API
  → transport/telegram/poller.go
  → allowlist check (identity)
  → idempotency (processed_updates)
  → Handler.HandleIncomingMessage(ctx, text)
  → ai.IntentResolver.Resolve(ctx, input)
  → app.UseCase.Execute(ctx, typedCommand)
  → infra.Repo + events.Publisher (same TX)
  → ResponseDTO
  → transport/telegram/formatter.go (HTML)
  → Telegram API sendMessage
```

HTTP path (`transport/http/api`) вызывает **те же** use cases — единая бизнес-логика для Telegram и Mini App.

---

## 8. Scheduler Flow

```
platform/scheduler (ticker 1m)
  → SELECT scheduled_jobs WHERE run_at <= now AND status='pending'
    FOR UPDATE SKIP LOCKED
  → switch job_type:
      reminder        → DeliverReminder → TelegramNotifier
      morning_review  → query.MorningReview → TelegramNotifier
      evening_review  → query.EveningReview → TelegramNotifier
  → UPDATE status=done (or reschedule, retry)
  → check quiet hours (settings)
```

Единая таблица `scheduled_jobs` ([ADR-007](../adr/007-builtin-scheduler.md)).  
Отдельной execution-таблицы `reminders` нет.

---

## 9. Event Log

```
Use Case (after successful TX)
  → events.Publisher.Publish(DomainEvent)
  → INSERT domain_events (aggregate_type, aggregate_id, event_type, payload, source)
```

Append-only, не full event sourcing ([ADR-006](../adr/006-domain-event-log.md)).

---

## 10. Ports & Adapters

| Port | Interface location | Adapter v1 | Future |
|------|-------------------|------------|--------|
| TaskRepository | tasks/domain or infra | SQLC/Postgres | same |
| IntentResolver | ai/ports | RuleBased first; optional Ollama on unknown (LIFEOS_LLM_ENABLED) | API LLM |
| Assistant | ai/ports | Template; optional Ollama with HTML-safe `<b>` + template fallback | — |
| Notifier | notifications/app | Telegram | Email, Push |
| Clock | platform/clock | System | Fake (tests) |
| ProjectReader / SphereReader | consumer `app` ports | projects/spheres infra | same |
| TransactionManager | platform/postgres | pgx TX | same |

---

## 11. Database

| | |
|---|---|
| Engine | PostgreSQL 16 |
| Schema | Single `public` schema |
| Ownership | Таблица принадлежит одному модулю |
| Cross-FK | M:N через junction: `task_projects`, `project_spheres` |
| PK | UUID v7 (time-sortable) |
| Money (Phase 2) | `amount_cents int64` |
| Timestamps | UTC in DB, display in user TZ |

### Core Tables

`users`, `user_settings`, `tasks`, `task_projects`, `projects`, `project_spheres`, `life_spheres`, `day_availability`, `scheduled_jobs`, `domain_events`, `processed_updates`, `finance_*`, `debts`, `habits`, `habit_logs`, `calendar_events`, `notes`, `note_tags`, `health_*`, `career_*`

---

## 12. API

| Method | Path | Auth |
|--------|------|------|
| GET | `/health`, `/ready`, `/metrics` | none |
| POST | `/api/v1/auth/token` | API key (`X-API-Key`) → JWT по `telegram_id` |
| POST | `/api/v1/auth/telegram-webapp` | Telegram Mini App `initData` → JWT |
| * | `/api/v1/*` | JWT Bearer |

**Mini App auth:** клиент шлёт `POST /api/v1/auth/telegram-webapp` с телом `{ "init_data": "<Telegram.WebApp.initData>" }`.
Сервер проверяет HMAC-SHA256 (`secret_key = HMAC_SHA256(bot_token, key=WebAppData)`), отклоняет просроченный `auth_date` (TTL = `LIFEOS_JWT_TTL_HOURS`, по умолчанию 24h), делает `EnsureUserByTelegram` и выдаёт JWT.

OpenAPI: [docs/api/openapi.yaml](../api/openapi.yaml).  
Webhook `/webhook/telegram` — optional (см. `LIFEOS_TELEGRAM_MODE=webhook`).

---

## 13. Stack

### MVP (default compose)

| Layer | Technology |
|-------|------------|
| Language | Go 1.25 |
| HTTP router | Chi |
| DB driver | pgx/v5 |
| Queries | SQLC |
| Migrations | Goose |
| CLI | Cobra |
| Telegram | Long polling |
| Logging | slog (JSON prod, text dev) |
| Metrics | Prometheus |
| Testing | testcontainers-go |
| CI | GitHub Actions |
| Lint | golangci-lint + depguard (import rules) |
| Dev reload | Air (optional) |

### Compose Profiles

```bash
docker compose up                                    # app + postgres
docker compose --profile observability up            # + prometheus + grafana
make observability-up                                # same profile: prometheus, grafana, jaeger
docker compose --profile cache up                    # + redis (Phase 2)
```

**Metrics:** Chi mounts `GET /metrics` (`promhttp`); Prometheus scrape target is `app:8080` in `deployments/prometheus/prometheus.yml`. Tracing via OTel is optional (`LIFEOS_OTEL_ENABLED`).

### Deferred from MVP default compose

Redis (profile `cache`), Grafana/Jaeger (profile `observability`), Telegram webhook — optional.  
JWT + REST + Mini App **уже в коде** (добавлены после исходного MVP scope; см. revision note в [ADR-009](../adr/009-mvp-infra-scope.md)).

---

## 14. Конфигурация

```bash
LIFEOS_DATABASE_URL=postgres://...
LIFEOS_HTTP_ADDR=:8080
LIFEOS_LOG_LEVEL=info
LIFEOS_LOG_FORMAT=json          # json|text

TELEGRAM_BOT_TOKEN=
LIFEOS_TELEGRAM_MODE=polling    # polling|webhook
LIFEOS_JWT_SECRET=
LIFEOS_JWT_TTL_HOURS=168        # Mini App session TTL
LIFEOS_API_KEY=
LIFEOS_MINIAPP_URL=             # public HTTPS …/app/ for web_app button
LIFEOS_STATIC_DIR=web/miniapp/dist
LIFEOS_LLM_ENABLED=false        # optional Ollama (rule-based first; degrade on down)
LIFEOS_OTEL_ENABLED=false
```

---

## 15. Локальный деплой

```
┌──────────┐     long polling      ┌─────────────┐
│ Telegram │◀────────────────────▶│  lifeos-app  │
│   API    │                       │  (Docker)   │
└──────────┘                       └──────┬──────┘
                                          │
                                   ┌──────▼──────┐
                                   │ PostgreSQL  │
                                   │  (volume)   │
                                   └─────────────┘
```

Polling — без port forward, TLS, static IP ([ADR-004](../adr/004-telegram-polling-mvp.md)).  
Webhook + static IP — Phase 1.5.

---

## 16. Тестирование

| Layer | What | How |
|-------|------|-----|
| domain | invariants, rules | unit, table-driven |
| app | use cases | unit, manual mock repos |
| ai | intent parsing | golden files (`testdata/`) |
| infra | SQLC repos | testcontainers PG |
| transport | webhook handler | httptest |
| e2e | create task flow | testcontainers + fake TG |
| scheduler | fire at T | inject FakeClock |

Coverage gate: ≥80% on `domain` + `app` packages only.

---

## 17. Observability (MVP → Hardening)

| Component | MVP | Sprint 7 |
|-----------|-----|----------|
| slog + request_id | ✅ | |
| Prometheus /metrics | ✅ | |
| Grafana dashboard | profile | ✅ |
| OpenTelemetry traces | | ✅ |

---

## 18. Безопасность (summary)

| Control | MVP |
|---------|-----|
| Telegram allowlist | ✅ P0 |
| Secrets in env | ✅ |
| SQL injection | SQLC only |
| Log token redaction | ✅ |
| JWT | ✅ REST + Mini App |
| Webhook secret | ✅ optional |

---

## 19. C4 — Container Diagram

```
┌──────────────────────────────────────────────────────────┐
│                     LifeOS Process                        │
│  ┌────────────┐  ┌─────────────┐  ┌──────────────────┐  │
│  │  Telegram  │  │  HTTP/Chi   │  │    Scheduler     │  │
│  │  Transport │  │  Transport  │  │    Worker        │  │
│  └─────┬──────┘  └──────┬──────┘  └────────┬─────────┘  │
│        └────────────────┼───────────────────┘            │
│                         ▼                                │
│              ┌─────────────────────┐                     │
│              │   Application Layer  │                     │
│              │   + Query Services   │                     │
│              └──────────┬──────────┘                     │
│                         ▼                                │
│              ┌─────────────────────┐                     │
│              │   Domain Layer       │                     │
│              └──────────┬──────────┘                     │
│                         ▼                                │
│              ┌─────────────────────┐                     │
│              │   Infrastructure     │                     │
│              └─────────────────────┘                     │
└──────────────────────────┬───────────────────────────────┘
                           │
                           ▼
                     PostgreSQL
              (Redis — Phase 2 profile)
                           │
                           ▼
                     Telegram API
```

---

## 20. Enforcement

- `depguard` / `go-arch-lint`: запрет cross-import `internal/*/domain`
- Code review: transport не содержит if/else бизнес-логики
- ADR required для отклонений от этого документа

---

## 21. Document Index

| Doc | Purpose |
|-----|---------|
| [DOMAIN_MODEL.md](DOMAIN_MODEL.md) | Entity detail |
| [../diagrams/ER.md](../diagrams/ER.md) | Schema |
| [../diagrams/SEQUENCE.md](../diagrams/SEQUENCE.md) | Flows |
| [../adr/](../adr/) | Decisions |
| [../roadmap/ROADMAP.md](../roadmap/ROADMAP.md) | Roadmap |
| [../api/openapi.yaml](../api/openapi.yaml) | REST contract |
| [../agents/](../agents/) | Orchestrator + backend / frontend / telegram agents |
