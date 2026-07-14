# TASK-003: Stage 2.1 — API support for Mini App daily cycle

- **Agent:** backend
- **Status:** DONE
- **Priority:** P1
- **Stage:** miniapp-depth (Stage 2.1)

## Goal

Закрыть backend-дыры, которые блокируют Habits / Calendar / Settings в Mini App: DTO/JSON shape, missing fields, 400/404 semantics, OpenAPI parity.

## In scope

- Habits / calendar / settings (/spheres) HTTP handlers + app layer as needed
- Align response shapes with `web/miniapp/src/api/types.ts` expectations
- OpenAPI sync for touched routes
- Tests for changed endpoints

## Out of scope

- Mini App React
- Telegram UX / keyboards

## Focus checklist

- `GET /habits/today` fields: `id, name, today_completed, streak`
- `POST /habits`, `POST /habits/{id}/track`
- `GET /calendar/today`, `POST /calendar/events` (`starts_at` ISO)
- `GET /settings` + morning/evening/quiet-hours
- spheres CRUD under `/settings/spheres`

## Acceptance

- [x] Contract matches Mini App or report precise mismatch for Frontend
- [x] Tests green for touched packages
- [x] Report: `docs/agents/reports/backend/TASK-003.md`
- [x] Commit + push branch
