# ADR-009: MVP Infrastructure Scope

## Status
Accepted

## Context
Аудит спецификации выявил преждевременную сложность: Redis, JWT, OpenTelemetry и Grafana в default Docker Compose для single-user local bot. Idempotency реализуема через PostgreSQL (`processed_updates`). JWT не нужен без REST API.

## Decision
**MVP default compose:** только `app` + `postgres`.

| Component | MVP default | When added |
|-----------|---------------|------------|
| PostgreSQL | ✅ | Sprint 0 |
| Redis | ❌ (compose profile `cache`) | Phase 2 / multi-instance |
| JWT | ❌ | Phase 3 (REST API) |
| OpenTelemetry | ❌ | Sprint 7 (hardening) |
| Grafana | ❌ (compose profile `observability`) | Sprint 7 |
| Prometheus | ✅ `/metrics` endpoint | Sprint 1 |
| Webhook | ❌ | Phase 1.5 |

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
