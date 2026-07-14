# LifeOS

Персональная операционная система для управления задачами, проектами, финансами, привычками и здоровьем.

**Telegram — основной UI. REST API + Telegram Mini App — второй клиент. Бизнес-логика не зависит от транспорта.**

## Статус

**Phase 2 (M7+)** — finance, habits, calendar, projects/spheres, health, career, knowledge, REST API, Mini App scaffold.

## Быстрый старт

```bash
cp .env.example .env
# заполнить TELEGRAM_BOT_TOKEN и при необходимости LIFEOS_JWT_SECRET / LIFEOS_API_KEY

make docker-up      # postgres + app
make migrate-up     # или: go run ./cmd/lifeos migrate up
make dev            # локально без docker
```

Mini App (dev):

```bash
cd web/miniapp && npm install && npm run dev
```

Mini App через Telegram (нужен публичный HTTPS):

```bash
make stack-up
# поднимает backend+frontend в Docker и HTTPS-туннель (cloudflared)
# URL пишется в LIFEOS_MINIAPP_URL; в боте появляется кнопка «📱 Mini App»
```

Либо вручную:

```bash
make miniapp-build
make docker-up
make tunnel          # scripts/https-tunnel.sh → обновляет .env
make docker-up       # пересоздать app с новым LIFEOS_MINIAPP_URL
```

## Документация

| Раздел | Путь |
|--------|------|
| Архитектура | [docs/architecture/ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md) |
| Domain Model | [docs/architecture/DOMAIN_MODEL.md](docs/architecture/DOMAIN_MODEL.md) |
| ER / Sequence | [docs/diagrams/](docs/diagrams/) |
| ADR (001–009) | [docs/adr/](docs/adr/) |
| Roadmap | [docs/roadmap/ROADMAP.md](docs/roadmap/ROADMAP.md) |
| Mini App UX/UI | [docs/miniapp/UX_UI_PLAN.md](docs/miniapp/UX_UI_PLAN.md) |
| Mini App local | [docs/miniapp/LOCAL_DEV.md](docs/miniapp/LOCAL_DEV.md) |
| Mini App Frontend lead | [docs/miniapp/FRONTEND_LEAD_PROMPT.md](docs/miniapp/FRONTEND_LEAD_PROMPT.md) |
| Mini App Backend | [docs/miniapp/BACKEND_PROMPT.md](docs/miniapp/BACKEND_PROMPT.md) |
| OpenAPI | [docs/api/openapi.yaml](docs/api/openapi.yaml) |
| Агенты (Architect / Backend / Frontend / Telegram) | [docs/agents/](docs/agents/) |

## Деплой

Local Mac, Docker Compose: **app + postgres** (default).

```bash
docker compose -f deployments/docker-compose.yml up
docker compose -f deployments/docker-compose.yml --profile observability up   # + prometheus, grafana, jaeger
# or: make observability-up
```

`GET /metrics` exposes Prometheus metrics on the app (scraped by `deployments/prometheus/prometheus.yml` → `app:8080`). OTel tracing is off by default (`LIFEOS_OTEL_ENABLED=false`).

Telegram: **long polling** по умолчанию; webhook — опционально.

## Стек

**Backend:** Go 1.25 · Chi · PostgreSQL 16 · pgx/v5 · SQLC · Goose · Cobra · slog · Prometheus · JWT

**Frontend:** React 19 · Vite · Tailwind CSS 4 · Telegram WebApp SDK

**Optional:** Ollama (LLM intent resolver) · OpenTelemetry

## Make targets

| Target | Описание |
|--------|----------|
| `make dev` | Запуск сервера |
| `make test` | Unit-тесты + coverage |
| `make coverage-check` | Gate: critical domain/app packages vs minima |
| `make test-integration` | HTTP API + e2e |
| `make lint` | golangci-lint |
| `make migrate-up` | Goose migrations |
| `make docker-up` | Compose build + start |
| `make observability-up` | Prometheus + Grafana + Jaeger (compose profile) |
| `make ci` | tidy + lint + test + coverage-check + build |
