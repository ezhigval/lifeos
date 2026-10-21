# LifeOS — Architecture (единый лист)

**Version:** 0.5  
**Status:** Current (synced to code, 2026-10-22)  
**Detail:** [DOMAIN_MODEL.md](DOMAIN_MODEL.md), [ROADMAP.md](ROADMAP.md)

---

## 1. Что это

| | |
|---|---|
| **Стиль** | Modular Monolith |
| **Deploy** | Один binary (`lifeos`), один процесс; отдельно — тонкий десктоп-клиент (`lifeos-desktop`) |
| **Пользователи v1** | 1 (single-tenant, `user_id` на всех таблицах) |
| **UI v1** | Telegram (long polling) + REST `/api/v1` + Telegram Mini App + десктоп-окно (Mac/Win) |
| **Хранилище** | PostgreSQL 16 (source of truth); SQLite только как HTTP-кэш десктопа |
| **LLM** | Опционален: OpenAI-совместимый (Groq по умолчанию) или Ollama; rule-based работает без него |
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
              ┌────────────┴────────────┐
              ▼                         ▼
       ┌─────────────┐          ┌─────────────┐
       │   memory    │          │  learning   │  ← память агента, opt-in обучение
       └─────────────┘          └─────────────┘
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
| identity | `internal/identity/` | User, логин-код вне Telegram (`login_code`) |
| settings | `internal/settings/` | UserSettings (+ флаги памяти/обучения, виджеты home) |
| tasks | `internal/tasks/` | Task (M:N projects via `task_projects`, M:N spheres, правила) |
| projects | `internal/projects/` | Project (M:N spheres) |
| spheres | `internal/spheres/` | LifeSphere |
| planning | `internal/planning/` | DayAvailability |
| finance | `internal/finance/` | Transaction, Debt, Category, PlannedCashflow |
| habits | `internal/habits/` | Habit, HabitLog |
| calendar | `internal/calendar/` | CalendarEvent |
| knowledge | `internal/knowledge/` | Note (связь с целью/note_target_link) |
| health | `internal/health/` | Weight, Steps, Sleep |
| career | `internal/career/` | Contact, Skill |
| notifications | `internal/notifications/` | ScheduledJob (нет `domain/` — логика в app/infra) |
| memory | `internal/memory/` | Memory (личная память агента, `user_memories`) |
| learning | `internal/learning/` | Event (анонимные события обучения, opt-in) |
| ai | `internal/ai/` | Ports + адаптеры (см. ниже) |
| query | `internal/query/` | Read services: priorities, reviews, analytics |

Внутри `internal/ai/`: `rulebased`, `openaiapi`, `ollama`, `composite`,
`templateassistant`, `agent` (tools-цикл диалога), `dialogue` (сервис чата),
`stt` (голос), `vision` (фото), `reviewsafe` (санитайзер HTML).

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
│       ├── api.go                  # HTTP API deps assembly
│       ├── agent_wire.go           # сборка агента: tools, память, обучение
│       └── stt_wire.go, vision_wire.go  # опциональные голос/фото
├── cmd/lifeos-desktop/             # тонкий десктоп-клиент (UI + SQLite-кэш + прокси к API)
├── internal/
│   ├── platform/                   # config, postgres, auth, scheduler, events, ids, db(sqlc), otel, logging
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
│   ├── memory/     {domain, app, infra}   # личная память агента
│   ├── learning/   {domain, app, infra}   # анонимные события обучения (opt-in)
│   ├── notifications/ {app, infra}        # ScheduledJob; deliver → TelegramNotifier
│   ├── ai/         {rulebased, openaiapi, ollama, composite, templateassistant,
│   │                agent, dialogue, stt, vision, reviewsafe}
│   ├── query/                      # priorities, reviews, analytics
│   └── transport/
│       ├── http/                   # Chi server + /api/v1 (JWT) + static Mini App
│       └── telegram/               # polling/webhook, screens, keyboards, FSM, media
├── web/miniapp/                    # Telegram Mini App (React 19 + Vite) + desktop UI variant
├── desktop/macos/                  # Swift .app оболочка (Mac-окно)
├── packaging/desktop-win/          # LifeOS.bat / install.bat для Windows
├── migrations/                     # Goose SQL (00001–00037)
├── queries/                        # SQLC sources
├── deployments/                    # Dockerfile, compose, prometheus/grafana, systemd, cloudflared, vps
├── scripts/                        # build-mac-app, build-win-desktop, package, backup/restore, проверка секретов
├── e2e/
├── docs/                           # architecture, ADR, deploy, api, ai, roadmap
├── Makefile
├── sqlc.yaml
└── .github/workflows/              # ci.yml, desktop-release.yml
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
intentResolver := ai.NewRuleBasedResolver() // + optional LLM composite (openaiapi / ollama)
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
| IntentResolver | ai/ports | RuleBased first; optional LLM on unknown (`LIFEOS_LLM_ENABLED`) — OpenAI-compatible (`openaiapi`) or Ollama, via `composite` | — |
| Assistant | ai/ports | Template; optional LLM (OpenAI-compatible / Ollama) with HTML-safe `<b>` + template fallback (`reviewsafe`) | — |
| ConversationalAgent | ai/ports | `agent` (tools-цикл: задачи, финансы, привычки, календарь и др.) + `dialogue.Service`; память (`memory`), обучение (`learning`) | — |
| SpeechToText | ai/ports | OpenAI-совместимый whisper (Groq default), `LIFEOS_STT_ENABLED` | локальные модели |
| Vision | ai/ports | OpenAI-совместимый vision (Groq Llama default), `LIFEOS_VISION_ENABLED` | — |
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

`users`, `user_settings`, `tasks`, `task_projects`, `task_spheres`, `projects`, `project_spheres`, `life_spheres`, `sphere_domain_links`, `day_availability`, `scheduled_jobs`, `domain_events`, `processed_updates`, `telegram_sessions`, `login_codes`, `finance_categories`, `finance_transactions`, `planned_cashflows`, `debts`, `habits`, `habit_logs`, `calendar_events`, `notes`, `note_tags`, `health_weight_logs`, `health_step_logs`, `health_sleep_logs`, `career_contacts`, `career_skills`, `user_memories`, `anon_learning_events` (миграции 00001–00037; `goals` удалены в 00024).

---

## 12. API

| Method | Path | Auth |
|--------|------|------|
| GET | `/health`, `/ready`, `/metrics` | none |
| POST | `/api/v1/auth/token` | API key (`X-API-Key`) → JWT по `telegram_id` |
| POST | `/api/v1/auth/telegram-webapp` | Telegram Mini App `initData` → JWT |
| POST | `/api/v1/auth/telegram-login/*` | код из чата бота → JWT (вход вне Telegram) |
| POST | `/api/v1/assistant/chat` | JWT Bearer (тот же диалоговый агент, что в боте) |
| * | `/api/v1/*` | JWT Bearer |

**Mini App auth:** клиент шлёт `POST /api/v1/auth/telegram-webapp` с телом `{ "init_data": "<Telegram.WebApp.initData>" }`.
Сервер проверяет HMAC-SHA256 (`secret_key = HMAC_SHA256(bot_token, key=WebAppData)`), отклоняет просроченный `auth_date` (TTL = `LIFEOS_WEBAPP_AUTH_TTL_HOURS` / `LIFEOS_JWT_TTL_HOURS`), делает `EnsureUserByTelegram` и выдаёт JWT.
Вне Telegram (десктоп, браузер) — одноразовый код из чата бота (`telegram_login_codes`).

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
LIFEOS_LLM_ENABLED=false        # optional LLM: openai (default Groq) or ollama; see docs/ops/LLM.md
LIFEOS_LLM_AGENT_ENABLED=false  # диалоговый агент с tools (нужен включённый LLM)
LIFEOS_STT_ENABLED=false        # распознавание голосовых сообщений
LIFEOS_VISION_ENABLED=false     # разбор фотографий
LIFEOS_OTEL_ENABLED=false
```

Полный список переменных — в [.env.example](../../.env.example); описание LLM-настроек — [docs/ops/LLM.md](../ops/LLM.md).

### Десктоп-переменные

```bash
LIFEOS_API_ORIGIN=https://…   # сервер, к которому проксирует lifeos-desktop
LIFEOS_DESKTOP_ADDR=127.0.0.1:0  # локальный адрес окна (0 = случайный порт)
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
Webhook + публичный HTTPS — опционально (`LIFEOS_TELEGRAM_MODE=webhook`, см. [docs/deploy/EDGE.md](../deploy/EDGE.md)).

### Десктоп (Mac / Windows)

```
┌───────────────────────┐    HTTPS     ┌───────────────────────┐
│ lifeos-desktop        │────────────▶ │ удалённый lifeos-app  │
│  UI (embedded)        │              │ (ВМ / ноутбук)        │
│  SQLite cache.db      │◀──────────── │ PostgreSQL            │
│  браузер-оболочка     │ proxy+cache  └───────────────────────┘
└───────────────────────┘
```

`cmd/lifeos-desktop` — тонкий клиент: раздаёт встроенный UI, кэширует ответы API в локальный SQLite (`modernc.org/sqlite`), проксирует `/api/*` и чат агента на `LIFEOS_API_ORIGIN`. Бизнес-логики на клиенте нет; вход — одноразовый код из чата бота. Карта клиентов: [docs/roadmap/CLIENTS.md](../roadmap/CLIENTS.md).

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
| Learning opt-in, HMAC anon subject, no PII in events | ✅ |

Политика раскрытия уязвимостей и чек-лист деплоя: [../../SECURITY.md](../../SECURITY.md).

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

Внешние клиенты одного процесса:
  Telegram app ─ polling/webhook ─▶ transport/telegram
  Mini App (React) ─ HTTPS /api/v1 + /app/ ─▶ transport/http
  lifeos-desktop (Mac/Win) ─ HTTPS + SQLite-кэш ─▶ transport/http (тонкий клиент)
  Ollama / Groq (опционально) ◀─ LLM/STT/Vision ─▶ internal/ai
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
| [../api/openapi.yaml](../api/openapi.yaml) | REST contract |
| [../getting-started.md](../getting-started.md) | Local and Docker run |
| [../deploy/vm.md](../deploy/vm.md) | VM, autodeploy, Telegram tunnel |
| [ROADMAP.md](ROADMAP.md) | Дорожная карта и статус срезов |
| [../../SECURITY.md](../../SECURITY.md) | Security policy |
| [../../CHANGELOG.md](../../CHANGELOG.md) | Релизы (Keep a Changelog + SemVer) |
| [../roadmap/CLIENTS.md](../roadmap/CLIENTS.md) | Клиенты: бот, Mini App, десктоп |
