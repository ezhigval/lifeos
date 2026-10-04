# Backlog

**Version:** 0.4  
**Synced to code:** 2026-07-14  
**See also:** [ROADMAP.md](ROADMAP.md) · [agents/](../agents/)

Приоритет: **P0** (blocker) · **P1** (high) · **P2** (medium) · **P3** (later)

Статусы: ✅ done · 🚧 partial · ⏳ open · 🧊 icebox

> Исторические sprint stories ниже помечены статусом относительно кода на main.  
> Новые задания агентам пишутся в `docs/agents/inbox/`, не копируются слепо отсюда.

---

## Epic: Infrastructure

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| INF-01…16 | Skeleton, compose, goose, sqlc, cobra, slog, metrics, CI, lint, Air, Makefile, testcontainers, Redis profile, RUNBOOK | P0–P3 | ✅ (OTel/Grafana/Redis — optional profiles) |

## Epic: Identity

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| ID-01 | users table + migration | P0 | ✅ |
| ID-02 | Telegram user allowlist | P0 | ✅ |
| ID-03 | Seed default user | P0 | ✅ |
| ID-04 | JWT generator/validator | P1 | ✅ |
| ID-05 | Mini App initData → JWT | P1 | ✅ |

## Epic: Settings

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| ST-01 | user_settings table | P1 | ✅ |
| ST-02 | Default review times | P1 | ✅ |
| ST-03 | Timezone-aware time parsing | P1 | ✅ |
| ST-04 | Quiet hours via API/bot | P2 | ✅ |

## Epic: Tasks

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| TK-01…08 | Core task use cases + tests | P0–P1 | ✅ |
| TK-09 | CancelTask + EditTask | P0 | ✅ merged (task-lifecycle) |
| TK-10 | duration_minutes + tags (hashtags) | P0 | ✅ merged |
| TK-11 | Reschedule single task (due_date persist) | P0 | ✅ merged |
| TK-12 | Auto-reschedule incomplete at evening review + notify | P1 | ✅ merged |
| TK-13 | Filter tasks by hashtag (API/Telegram) | P1 | ✅ merged |
| TK-14 | Mini App task detail/edit UX | P1 | 🚧 via miniapp-ux merge |

## Epic: Goals → Projects (superseded)

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| GL-* | goals table / CreateGoal / progress | — | ✅ superseded by `projects` (00022–00024) |

## Epic: Projects & Spheres

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| PJ-01 | projects + project_spheres | P1 | ✅ |
| PJ-02 | life_spheres CRUD (API + bot drafts) | P1 | ✅ |
| PJ-03 | Mini App sphere tree UX | P2 | 🚧 Home/Spheres есть; Settings CRUD — WIP |

## Epic: Planning / Notifications / AI / Events / Query

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| PL / NT / AI / EV / QR core | As in Phase 1–2 | P0–P1 | ✅ |
| AI-06 | LLM adapter production-ready | P2 | 🚧 Ollama optional |
| TG-07 | Webhook mode | P3 | ✅ code optional |
| TG-08 | Thin handler refactor (strangler) | P2 | ⏳ |
| TG-09 | Shared presentation extraction | P3 | ⏳ |

## Epic: Finance / Habits / Calendar / Domains

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| FN / HB / CAL / knowledge / health / career | Telegram + REST | P2 | ✅ |
| MA-FIN | Mini App finance overview/ring | P2 | 🚧 client scaffold; overview API gaps TBD |
| MA-HAB / MA-CAL | Mini App habits/calendar screens | P2 | 🚧 WIP branches |
| MA-SET | Mini App settings + spheres CRUD UI | P2 | ⏳ |

## Epic: Workspaces (shared spheres) — Stage 4, после фикса багов

| ID | Story | Priority | Status |
|----|-------|----------|--------|
| WS-01 | Миграции: workspaces + workspace_members + workspace_id во всех доменах (NULL=личное) | P0 | ⏳ |
| WS-02 | API: CRUD воркспейсов, инвайты сотрудников, роли owner/editor/viewer | P0 | ⏳ |
| WS-03 | Контекст запросов `?context=personal\|ws:<id>\|all` во всех list-API | P0 | ⏳ |
| WS-04 | Главная: экраны «Все / Личное / Воркспейсы»; свои задачи/события ws подмешиваются, чужие — нет; внутри ws — всё всех участников | P0 | ⏳ |
| WS-05 | Настройки: что шерить на главную (воркспейсы/домены) | P1 | ⏳ |
| WS-06 | Фин-синхронизация: «зарплата сотруднику X» в ws → личный планируемый доход X (идемпотентно, origin_txn_id) | P0 | ⏳ |
| WS-07 | Календарь воркспейса + общий вид на главной | P1 | ⏳ |
| WS-08 | Заметки / долги / привычки / аналитика в контексте ws | P1 | ⏳ |
| WS-09 | Telegram: `/ws <name>` смена контекста, создание в ws | P2 | ⏳ |
| WS-10 | Отдельная экономика ws: бюджеты/отчёты; опциональная сводка «личное+бизнес» | P2 | ⏳ |
| WS-11 | Multi-link задач ↔ несколько сфер/проектов (`task_spheres`; `task_projects` уже есть — нужен FE-мультивыбор и фильтры по всем связям) | P1 | ⏳ |
| WS-12 | Bidirectional notes/reminders + каскадный picker ws→сфера→проект→задача | P1 | ⏳ |
| WS-13 | Календарь в контексте ws, фильтры по воркспейсам/сферам (M/W/D — TASK-011 п.4) | P1 | ⏳ |
| WS-14 | Привычки в ws + подключение трекера к задаче (`habits.linked_task_id`) | P1 | ⏳ |
| WS-15 | Domain links сфер: деньги→финансы, здоровье→привычки, карьера→ws-мосты; автогенерация («зарплата Имя 100000» → planned расход + зеркало) | P0 | ⏳ |
| WS-16 | Настройки главной v2: `home.layout` — видимость/порядок блоков + какие ws подмешивать | P1 | ⏳ |

UX-правки owner (не ждут Stage 4): **TASK-011-ux-p0-fixes** — planned-формы валидация,
задачи на главной, habit CRUD+срок, календарь M/W/D, заметки/напоминания 2-way, настройки блоков.

Детали и модели: [ROADMAP.md § Stage 4](ROADMAP.md)

## Epic: Stabilization (next stage — bugfix)

| ID | Story | Owner agent | Priority | Status |
|----|-------|-------------|----------|--------|
| FIX-B01 | Audit P0 backend bugs in domain/HTTP only | Backend | P0 | ⏳ inbox |
| FIX-T01 | Audit P0 Telegram UX/FSM/keyboard bugs | Telegram | P0 | ⏳ inbox |
| FIX-F01 | Audit P0 Mini App auth/nav/render bugs | Frontend | P0 | ⏳ inbox |

Конкретные баги заполняются Architect в `docs/agents/inbox/` после joint planning / dogfood list.

---

## Mini App (active)

План: [UX_UI_PLAN.md](../miniapp/UX_UI_PLAN.md)

| ID | Story | Priority | Phase |
|----|-------|----------|-------|
| MA-A1 | Auth `POST /auth/telegram-webapp` | P0 | A (backend) |
| MA-A2 | Telegram BackButton on nested routes | P0 | A ✅ frontend |
| MA-A4 | Finance overview API + Mini App wire | P0 | A (backend) |
| MA-A3 | Tab «Ещё» + Settings stub | P1 | A ✅ frontend |
| MA-A5 | Sheet motion | P1 | A ✅ frontend |
| MA-A6 | Query error/retry | P1 | A ✅ frontend |
| MA-B4 | Habits today + track | P1 | B |
| MA-B1 | CreateTask sheet (priority/due) | P1 | B |

Остальные MA-* — в UX_UI_PLAN §8.

---

## Icebox

- Multi-user SaaS / family
- Standalone web app / mobile (после Mini App)
- Google Calendar sync
- Bank API integration
- GraphQL API
- ~~Voice messages → STT~~ (Telegram voice / video_note / audio via Whisper)
- ~~Vision for photos without caption~~ (Groq Llama 4 Scout → agent text)
- Mini App assistant chat UI (API exists, UI deferred)
- i18n framework (пока RU-only константы)
