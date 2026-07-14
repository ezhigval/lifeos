# Roadmap

**Version:** 0.4  
**Synced to code:** 2026-07-14  
**See also:** [ARCHITECTURE.md](../architecture/ARCHITECTURE.md) · [agents/CURRENT_STATE.md](../agents/CURRENT_STATE.md)

> Следующий ребилд фаз и багфикс-спринт — совместно с Architect после sync документации.  
> Этот файл отражает **факт кода на main**, не желаемое будущее без подтверждения.

---

## Horizon Overview

```
2026 Q2–Q3              Now                     Next (планируем)
──────────────          ─────────               ────────────────
Foundation ✅           Expansion ✅            Stabilization →
(M1–M4 done,            Domains + REST +        Bugfix (P0)
 M5 dogfood 🚧)         Mini App scaffold       Intelligence polish
                                                Mini App depth
                                                Multi-client later
```

---

## Phase 0: Design ✅

- [x] Domain Model, Architecture, ADRs
- [x] ER + Sequence diagrams
- [x] Roadmap, Backlog, Sprint Plan (historical)

---

## Phase 1: Foundation (MVP) ✅

**Goal:** Telegram бот — capture, tasks, reminders, reviews.

| Milestone | Status |
|-----------|--------|
| M1 Skeleton | ✅ |
| M2 Core Domain | ✅ |
| M3 Telegram + NL | ✅ |
| M4 Scheduler + Reviews | ✅ |
| M5 Hardening | 🚧 OTel/backup есть; 14-day dogfooding gate открыт |

---

## Phase 2: Expansion ✅ (core)

| Milestone | Status |
|-----------|--------|
| M6 Finance | ✅ |
| M7 Habits + Calendar + Projects/Spheres | ✅ (goals → projects) |
| M8 Analytics | ✅ query layer |

Также на main: knowledge, health, career, REST `/api/v1` + JWT, Mini App auth + Home/Spheres/Finance scaffold.

---

## Phase 3: Stabilization + Intelligence — **next planning focus**

Приоритет этапа 1 (согласовано владельцем): **bugfix по зонам агентов**, затем roadmap rebuild.

Кандидаты (не закоммичены как обязательства до совместного плана):

- [ ] P0/P1 bug sweep (Backend / Telegram / Frontend — каждый в своей зоне)
- [ ] OpenAPI ↔ router parity (ongoing)
- [ ] Thin Telegram handler / presentation boundaries
- [ ] LLM resolver production-ready (Ollama adapter есть)
- [ ] Assistant summaries polish for reviews
- [ ] Mini App: Habits / Calendar / Settings depth (ветка UX in flight)
- [ ] Task lifecycle gaps (duration/tags/edit — см. unmerged ветки)

---

## Phase 4: Clients + Life Domains (later)

- [ ] Mini App: не полный parity с ботом; browse/structure/visual vs NL capture
- [ ] Web / Mobile — **не сейчас** (явно отложено владельцем)
- [ ] Production module (orders, pipeline)
- [ ] Multi-user / family mode

---

## Technical Debt Budget

Каждый рабочий цикл: ~20% на tests, observability, docs, boundary refactoring.

---

## Decision Gates

| Gate | Criteria | Status |
|------|----------|--------|
| G0 → G1 (code) | Runnable app + CI | ✅ |
| G1 → G2 (daily use) | M1–M4 complete | ✅ |
| G2 → G3 (stable) | 14 days dogfooding | 🚧 |
| G3 → G4 (expansion) | Phase 2 domains | ✅ |
| G4 → G5 | Bugfix + agreed next roadmap | ⏳ planning |
