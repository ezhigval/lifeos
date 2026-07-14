# Current State (code audit)

**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Product:** LifeOS — personal life-ops modular monolith

---

## Decisions locked

- Bug list: free audit (no owner dogfood list)
- WIP merged: task-lifecycle + miniapp-ux
- Next after bugfix: **Mini App + functionality**

---

## Stack

| Layer | Tech |
|-------|------|
| Backend | Go 1.25 · Chi · pgx · SQLC · Goose · Cobra · slog · Prometheus · JWT |
| DB | PostgreSQL 16 (migrations through 00027 task duration/tags) |
| Telegram | Long polling default; webhook optional |
| Mini App | React 19 · Vite · Tailwind 4 · TanStack Query · BrowserRouter `/app` |
| Deploy | Docker Compose: `app` + `postgres`; profiles `observability`, `cache` |

---

## Recently merged

### Task lifecycle
Duration, tags, edit/cancel/reschedule, auto-reschedule incomplete, list-by-tag; HTTP + Telegram wiring.

### Mini App UX
More tab, Habits, Calendar, Settings, Task detail, Analytics/Notes/Health/Career/Debts/Reminders screens; BackButton; empty/error UI. Auth initData freeze retained from main.

---

## Agents (Stage 1)

TASK-001 DONE (backend) — free-audit bugfix in-zone; see `docs/agents/reports/backend/TASK-001.md`.
TASK-001 for frontend / telegram — track their inbox/reports.
See `docs/agents/inbox/*/TASK-001-bugfix-audit.md`.
