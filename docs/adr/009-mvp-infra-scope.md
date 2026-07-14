# ADR-009: MVP Infrastructure Scope

## Status
Accepted (historical MVP decision; see Revision below)

## Context
Аудит спецификации выявил преждевременную сложность: Redis, JWT, OpenTelemetry и Grafana в default Docker Compose для single-user local bot. Idempotency реализуема через PostgreSQL (`processed_updates`). JWT не нужен без REST API.

## Decision
**MVP default compose:** только `app` + `postgres`.

| Component | MVP default (decision) | Current code (2026-07) |
|-----------|------------------------|-------------------------|
| PostgreSQL | ✅ | ✅ |
| Redis | ❌ (compose profile `cache`) | profile `cache` |
| JWT | ❌ (defer to REST) | ✅ REST + Mini App |
| OpenTelemetry | ❌ | optional (`LIFEOS_OTEL_ENABLED`) |
| Grafana | ❌ (profile `observability`) | profile `observability` |
| Prometheus | ✅ `/metrics` | ✅ |
| Webhook | ❌ | optional (`LIFEOS_TELEGRAM_MODE=webhook`) |

## Revision (post-MVP)
Default compose по-прежнему `app` + `postgres`. JWT, REST `/api/v1`, Mini App static serving и OTel hooks добавлены как **опциональные/включаемые** возможности без раздувания default infra (нет Redis/Grafana в default).

## Consequences
**+** Проще `docker compose up`, меньше точек отказа  
**+** Быстрее Sprint 0–1  
**+** Idempotency без внешнего store  
**−** Multi-step Telegram dialogs defer до Phase 1.5 (PG state или Redis profile)  
**−** Distributed scheduler lock defer до multi-instance

## Alternatives Considered
- **Redis с Day 1** — отклонено: нет use case в single-instance
- **JWT skeleton в M2** — отклонено: мёртвый код без REST consumer
- **OTel с Sprint 1** — отклонено: slog + Prometheus достаточно для MVP
