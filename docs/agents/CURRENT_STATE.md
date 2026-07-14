# Current State (code audit)

**Date:** 2026-07-14  
**Ref:** `main` @ synced docs branch  
**Product:** LifeOS — personal life-ops modular monolith

---

## Stack (as running)

| Layer | Tech |
|-------|------|
| Backend | Go 1.25 · Chi · pgx · SQLC · Goose · Cobra · slog · Prometheus · JWT |
| DB | PostgreSQL 16 |
| Telegram | Long polling default; webhook optional |
| Mini App | React 19 · Vite · Tailwind 4 · TanStack Query · Router basename `/app` |
| Deploy | Docker Compose: `app` + `postgres`; profiles `observability`, `cache` |

---

## Backend modules present

`identity`, `settings`, `tasks`, `projects`, `spheres`, `planning`, `finance`, `habits`, `calendar`, `knowledge`, `health`, `career`, `notifications`, `query`, `ai`, `platform`, `transport/{http,telegram}`

**Goals** удалены → `projects` (+ `task_projects`, `project_spheres`). Migrations through `00026_telegram_session_states`.

Composition root: `cmd/lifeos/cmd/{serve,runtime,api}.go` (manual DI).

---

## REST `/api/v1` (implemented)

Public: `POST /auth/token`, `POST /auth/telegram-webapp`  
JWT: tasks, projects, reviews, priorities, analytics, finance, habits, reminders, notes, career, health, calendar, settings (+ **spheres** CRUD under `/settings/spheres`).

Contract: `docs/api/openapi.yaml` (kept in parity with `router.go`).

---

## Telegram surface

Commands: `/start`, `/cancel`, `/clear`, `/delete`  
Reply menu: Главная, Задачи, Привычки, Проекты, Календарь, Статистика, Настройки (+ Mini App web_app if URL set)  
Screen/dashboard edit-or-send + session FSM drafts  
IntentResolver: rule-based default, optional Ollama composite

---

## Mini App surface (main)

Routes: Home, Spheres tree → sphere → project  
Features: upcoming tasks + complete, finance card/sheet, TG auth session persist  
Nav: Главная + Сферы (Ещё / Habits / Calendar / Settings — depth incomplete on main)

---

## Docs vs code (after this sync)

| Item | Status |
|------|--------|
| ARCHITECTURE tree / composition (no goals) | fixed |
| SEQUENCE reviews/priorities use projects | fixed |
| OpenAPI spheres endpoints | fixed |
| ADR-002/009 outdated examples / JWT note | fixed |
| Roadmap/backlog Goals epic + JWT P3 | fixed |
| Agent roles Frontend + Telegram + Architect | added |

---

## Unmerged / WIP (not on main — plan carefully)

| Branch (remote) | Theme |
|-----------------|--------|
| `cursor/miniapp-ux-ui-plan-10dc` | Mini App Phase B/C screens, frontend prompts |
| `cursor/task-core-lifecycle-65c7` | Task duration/tags/lifecycle |
| `cursor/telegram-transport-agent-prompt-b841` | Telegram agent prompt (merged conceptually into `TELEGRAM.md`) |
