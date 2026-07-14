# Roadmap

**Version:** 0.3  
**See also:** [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)

---

## Horizon Overview

```
2026 Q3          2026 Q4          2027 Q1          2027+
────────         ────────         ────────         ────
Foundation ✅    Expansion ✅     Intelligence     Multi-client
(M1–M5)          (M6–M8)          (LLM polish)     (Mobile, sync)
```

---

## Phase 0: Design ✅

- [x] Domain Model, Architecture, ADRs
- [x] ER + Sequence diagrams
- [x] Roadmap, Backlog, Sprint Plan

---

## Phase 1: Foundation (MVP) ✅

**Goal:** Telegram бот — capture, tasks, reminders, reviews.

| Milestone | Status |
|-----------|--------|
| M1 Skeleton | ✅ |
| M2 Core Domain | ✅ |
| M3 Telegram + NL | ✅ |
| M4 Scheduler + Reviews | ✅ |
| M5 Hardening | 🚧 partial (OTel, backup scripts) |

**Exit criteria:** 14 days daily use, 0 P0 bugs — in progress.

---

## Phase 2: Expansion ✅ (mostly)

| Milestone | Status |
|-----------|--------|
| M6 Finance | ✅ |
| M7 Habits + Calendar + Projects/Spheres | ✅ |
| M8 Analytics | ✅ query layer |

Дополнительно реализовано: **knowledge, health, career**, REST API + JWT, Mini App scaffold.

---

## Phase 3: Intelligence — next

- [ ] LLM resolver production-ready (Ollama adapter exists)
- [ ] Assistant summaries for reviews
- [ ] Webhook Telegram (optional, code exists)
- [ ] OpenAPI sync with all endpoints

---

## Phase 4: Life Domains — ongoing

- [x] Knowledge, Career, Health (Telegram + API)
- [ ] Mini App: full feature parity with bot
- [ ] Production module (orders, pipeline)
- [ ] Multi-user / family mode

---

## Technical Debt Budget

Каждый sprint: ~20% на tests, observability, docs, boundary refactoring.

---

## Decision Gates

| Gate | Criteria | Status |
|------|----------|--------|
| G0 → G1 (code) | Runnable app + CI | ✅ |
| G1 → G2 (daily use) | M1–M4 complete | ✅ |
| G2 → G3 (stable) | 14 days dogfooding | 🚧 |
| G3 → G4 (expansion) | Phase 2 domains | ✅ |
