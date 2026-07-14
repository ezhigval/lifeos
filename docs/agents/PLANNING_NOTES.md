# Planning Notes — next session with owner

**Status:** READY FOR JOINT PLANNING  
**Prerequisite done:** documentation synced to code (this PR)

---

## Agreed constraints (owner)

1. Три агента: **Frontend** (Mini App only), **Backend** (logic/API), **Telegram** (bot transport/UX).
2. Стек: React+Vite · Go+PostgreSQL+Docker · modular / hexagonal boundaries.
3. Web / native mobile — **сильно позже**; сейчас Mini App.
4. После docs sync → **совместно** пересобрать roadmap.
5. **Этап 1 исполнения:** bug fix. Каждый агент — только своя зона.

---

## Proposed Phase map (draft — подтвердить)

```
Stage 0  ✅  Docs ↔ code sync + agent orchestration scaffold
Stage 1  →  Bugfix sweep (P0→P1) per agent zone
Stage 2  →  Stabilization debt (thin TG handler, OpenAPI parity CI, tests)
Stage 3  →  Mini App depth (Habits/Calendar/Settings) + needed API gaps
Stage 4  →  Intelligence polish (LLM, assistant)
Stage 5+ →  Multi-client / family / production module (icebox until then)
```

---

## Stage 1 — Bugfix (draft assignments)

Конкретный список багов владелец должен подтвердить (dogfood list). Пока — шаблоны аудита:

### Backend (`inbox/backend/TASK-001-bugfix-audit.md`)

- Регрессии API / domain / auth / finance-overview gaps
- Несоответствия DTO ↔ Mini App expectations
- Тесты на сломанные use cases

### Telegram (`inbox/telegram/TASK-001-bugfix-audit.md`)

- Reply keyboard persistence / screen edit failures
- FSM cancel / drafts / `/clear` `/delete`
- Notifier / dashboard regressions

### Frontend (`inbox/frontend/TASK-001-bugfix-audit.md`)

- Auth/session/initData edge cases
- Grey screen / routing / empty states
- Finance/tasks UI breakage vs live API

**Правило:** нашёл баг чужой зоны → файл в `reports/` + ping Architect, не фиксить самому.

---

## Open questions for owner (коротко)

1. Есть ли готовый список багов из личного dogfooding? Если да — вставляем в TASK-001 как P0.
2. Мержить ли WIP-ветки (`miniapp-ux`, `task-lifecycle`) **после** bugfix или выборочно?
3. Приоритет после Stage 1: Mini App depth или Intelligence?

Ответы → Architect обновляет ROADMAP + выдаёт финальные inbox tasks.
