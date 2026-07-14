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

1. ✅ Merge `cursor/task-core-lifecycle-65c7`
2. ✅ Merge `cursor/miniapp-ux-ui-plan-10dc`
3. ✅ TASK-001 bugfix DONE (all agents)
4. ✅ Cross-zone asks → Stage 2

## Stage 2 execution

1. ✅ Backend: `GET /api/v1/finance/overview?period=YYYY-MM`
2. ✅ Frontend: wire overview + polish
3. ✅ Telegram: UX align (keyboard/formatters)
4. Ongoing: Mini App depth / more API gaps as they surface
