# Roadmap

**Version:** 0.5  
**Synced:** 2026-07-14 (owner lock)  
**See also:** [ARCHITECTURE.md](../architecture/ARCHITECTURE.md) · [agents/PLANNING_NOTES.md](../agents/PLANNING_NOTES.md)

---

## Horizon

```
Now                         Next                        Later
────────                    ────────                    ─────
Stages 0–3 ✅               Real dogfood 14d            Web/Mobile icebox
Mini App + auth + LLM opt   OpenAPI CI / observ.
```

---

## Phase 0–2 (done)

Foundation (M1–M4) ✅ · Hardening dogfood 🚧 · Expansion domains + REST + Mini App scaffold ✅

---

## Stage 1 — Merge WIP + Bugfix (active)

- [x] Merge task lifecycle branch
- [x] Merge miniapp UX branch
- [x] Free-audit bugfix: Backend / Frontend / Telegram (TASK-001 DONE)
- [x] Cross-zone asks → Stage 2 OPEN

## Stage 2 — Mini App + functionality (active)

- [x] Backend TASK-002: `GET /finance/overview?period=YYYY-MM`
- [x] Frontend TASK-002: wire overview + polish
- [x] Telegram TASK-002: align UX with lifecycle / Mini App
- [x] **Stage 2.1 TASK-003:** Habits / Calendar / Settings daily cycle
- [x] **Stage 2.2 TASK-004:** Notes / Health / Career / Reminders / Debts / Analytics
- [ ] No full bot↔Mini App parity required

## Stage 3 — Stabilization + Intelligence ✅

- [x] **3.0 TASK-005:** Dogfood free-audit (Frontend / Backend / Telegram)
- [x] **3.1 TASK-006:** Thin Telegram handler (strangler) — handler.go ~−51%
- [x] **3.2 TASK-007:** Intelligence polish (LLM composite + assistant HTML-safe)

## Stage 4 — Shared Workspaces (planned, после закрытия багов TASK-008)

> Стартует **после** фикса P0-багов dogfood (`TASK-008-dogfood-p0`). Новая функциональность сейчас не добавляется.
> Задание для агентов: [docs/agents/inbox/TASK-010-workspaces.md](../agents/inbox/TASK-010-workspaces.md).

**Идея:** сфера (life_sphere) может быть шерирована как **воркспейс** (например «Бизнес»):
сотрудники, общая экономика, календарь, заметки, задачи — изолированно от личного,
но с возможностью «подмешивать» на главную и синхронизировать cross-context финансы.

### Модели данных (эскиз)
- `workspaces` — расширяет `life_spheres`: `is_shared`, `owner_id`, `slug`, `role` (owner/editor/viewer).
- `workspace_members` — user ↔ workspace + роль; сотрудники = участники воркспейса.
- Все доменные таблицы (tasks / events / transactions / notes / habits / debts …) получают
  nullable `workspace_id` (NULL = личное). Личное и воркспейсные данные не смешиваются в запросах по умолчанию.
- Экономическая связь: транзакция «зарплата → сотруднику X» в воркспейсе при участнике=X
  автоматически порождает зеркальную запись в личном контексте (планируемый доход). Идемпотентно, через `origin_txn_id`.

### Пользовательские истории
| ID | Story | Priority |
|----|-------|----------|
| WS-01 | Миграции: workspaces/members + `workspace_id` во всех доменах | P0 |
| WS-02 | API: CRUD воркспейсов, инвайты (Telegram-приглашение), роли | P0 |
| WS-03 | Контекст запросов: фильтр `?context=personal\|ws:<id>\|all` во всех list-API | P0 |
| WS-04 | Главная Mini App: экраны «Все / Личное / Воркспейсы», подмешивание задач и событий выбранных воркспейсов; чужие задачи сотрудников НЕ видны в «Всех» | P0 |
| WS-05 | Настройки: что шерить на главную (какие воркспейсы/домены), порядок блоков | P1 |
| WS-06 | Финансовая синхронизация: зарплата сотрудника→личный планируемый доход; двойные записи, идемпотентность | P0 |
| WS-07 | Календарь воркспейса + общий вид на главной (свои+выбранные ws); внутри воркспейса — дела/задачи всех участников | P1 |
| WS-08 | Заметки / долги / привычки / аналитика в контексте воркспейса | P1 |
| WS-09 | Telegram: смена контекста `/ws <name>`, создание задач/трат в воркспейсе | P2 |
| WS-10 | Разделение экономики: бюджеты/отчёты воркспейса отдельно, сводка «личное + бизнес» опционально | P2 |

### Decision Gate
| Gate | Criteria | Status |
|------|----------|--------|
| Bugfix → Stage 4 | TASK-008 P0 closed | ⏳ |

## Infra & Deploy (активно, параллельно с багфиксом)

- [x] Инструкция: заливка секретов на Yandex Cloud VM — [docs/deploy/YANDEX_CLOUD_SECRETS.md](../deploy/YANDEX_CLOUD_SECRETS.md) (без IP/данных)
- [x] `.gitignore` усилен: ключи (`*.key`, `id_rsa`), `secrets/`, локальные deploy-оверраиды
- [x] `deployments/docker-compose.override.example.yml` — шаблон prod-оверрая (реальный оверрай не коммитится)
- [ ] Owner: применить инструкцию на ВМ (YC), заменить dev-секреты на боевые

## Remaining debt (ongoing)

- [x] OpenAPI ↔ router parity CI (`make openapi-check`)
- [x] 14-day dogfood checklist: [DOGFOOD.md](DOGFOOD.md)
- [x] Test / observability budget (~20%) — coverage gates + TG regressions + `/metrics` note
- [ ] Owner runs 14-day dogfood gate (G2→G3) on Mac + Telegram
- [ ] Collect dogfood P0 → inbox `TASK-008-dogfood-p0` (DRAFT)

## Icebox

Web app · native mobile · family/multi-user · bank/calendar sync · GraphQL · STT
Mini App UX plan (detail): [docs/miniapp/UX_UI_PLAN.md](../miniapp/UX_UI_PLAN.md)

---

## Decision Gates

| Gate | Criteria | Status |
|------|----------|--------|
| G0 → G1 | Runnable + CI | ✅ |
| G1 → G2 | M1–M4 | ✅ |
| G2 → G3 | 14-day dogfood ([DOGFOOD.md](DOGFOOD.md)) | 🚧 |
| G3 → G4 | Phase 2 domains | ✅ |
| Stage1 → Stage2 | WIP merged + P0 bugs closed | ✅ |
| Stage2 → Stage3 | Mini App depth + contracts | ✅ |
| Stage3 complete | Dogfood audit + thin handler + LLM optional | ✅ |
