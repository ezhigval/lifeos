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

---

## Owner decisions (2026-10-04) — planning update

**Порядок работ:** сначала закрытие P0 (TASK-008 dogfood + TASK-011 UX-правки), затем Stage 4 Workspaces (TASK-010).

### TASK-011 (P0, сейчас) — расширен до 8 пунктов:
1. Planned income/expense — валидация формы (все обязательные поля видимы/понятны).
2. Главная не показывает задачи — фикс фильтров/query.
3. Привычки: edit/delete/срок (start/end).
4. Календарь: Month/Week/Day + фильтры по сферам (позже и ws).
5. Заметки/напоминания: двусторонняя связь с задачами + каскадный picker ws→сфера→проект→задача.
6. Настройки главной: тогглы блоков (home.layout).
7. Задачи ↔ несколько сфер/проектов (BE `task_projects` уже есть; добавить `task_spheres` + FE-мультивыбор).
8. Сфера «Деньги» → финансы: задача «зарплата Имя сумма» → planned-доход (domain_rule engine, personal-версия WS-15).

### TASK-010 (Stage 4, после P0) — добавлены WS-11..WS-16:
multi-link задач, bidirectional notes/reminders (polymorphic targets), календарь в контексте ws,
привычки+linked_task, domain links сфер (money→finance, health→habits, career→workspace-мосты:
проект «Бизнес» = вход в воркспейс без копирования данных), настройки главной v2 (+ какие ws подмешивать).

Детали: docs/roadmap/ROADMAP.md § Stage 4 · docs/roadmap/BACKLOG.md Epic WS · docs/agents/inbox/TASK-010-workspaces.md · docs/agents/inbox/TASK-011-ux-p0-fixes.md

**Deploy rules:** секреты никогда не коммитятся (см. docs/deploy/YANDEX_CLOUD_SECRETS.md); проверка перед push обязательна.
