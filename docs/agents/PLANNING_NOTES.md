# Planning Notes — locked

**Status:** LOCKED by owner 2026-07-14  
**Prerequisite:** docs sync PR + agent scaffold

---

## Owner answers

| # | Question | Answer |
|---|----------|--------|
| 1 | Dogfood bug list? | **Нет** → free audit P0/P1 в каждой зоне |
| 2 | WIP branches? | **Мержим** (`task-lifecycle`, `miniapp-ux`; TG agent prompt уже перекрыт `TELEGRAM.md`) |
| 3 | After Stage 1? | **Mini App + функционал** (не Intelligence first) |

---

## Phase map (confirmed)

```
Stage 0  ✅  Docs ↔ code + agent orchestration
Stage 1  →  Merge WIP → Bugfix free-audit (Backend / Frontend / Telegram)
Stage 2  →  Mini App depth + domain functionality (primary track)
Stage 3  →  Stabilization debt (thin handler, tests, OpenAPI CI) as needed
Stage 4  →  Intelligence polish (later)
Stage 5+ →  Web / mobile / family (icebox)
```

---

## Stage 1 execution

1. Merge `cursor/task-core-lifecycle-65c7` (tasks duration/tags/edit + TG/HTTP)
2. Merge `cursor/miniapp-ux-ui-plan-10dc` (Settings, Habits, Calendar, More, Phase C)
3. Open TASK-001 for all three agents (free audit)
4. Agents fix only their zone; cross-zone → report ask

**Правило:** нашёл баг чужой зоны → `reports/` + Architect, не фиксить самому.

---

## Stage 2 focus (после bugfix)

- Mini App: Habits / Calendar / Settings / More / Task detail — довести до рабочего UX
- Backend: API gaps под эти экраны (overview, mutate tasks, spheres settings)
- Telegram: совместимость с новым task lifecycle (форматтеры, клавиатуры, intents)
