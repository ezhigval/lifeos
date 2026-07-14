# Roadmap

**Version:** 0.5  
**Synced:** 2026-07-14 (owner lock)  
**See also:** [ARCHITECTURE.md](../architecture/ARCHITECTURE.md) · [agents/PLANNING_NOTES.md](../agents/PLANNING_NOTES.md)

---

## Horizon

```
Now                         Next                        Later
────────                    ────────                    ─────
Stage 0 docs ✅             Stage 2 Mini App +          Intelligence
Stage 1 merge+bugfix      functionality               Web/Mobile icebox
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

## Stage 3 — Stabilization + Intelligence (active)

- [x] **3.0 TASK-005:** Dogfood free-audit (Frontend / Backend / Telegram)
- [x] **3.1 TASK-006:** Thin Telegram handler (strangler)
- [x] **3.2 TASK-007:** Intelligence polish (LLM composite + assistant) — DONE

## Stage 3 — Stabilization (as needed)

- [ ] Thin Telegram handler (strangler)
- [ ] OpenAPI ↔ router parity checks
- [ ] Test / observability debt (~20% budget)

## Stage 4 — Intelligence (later)

- [ ] LLM resolver production-ready
- [ ] Assistant summaries polish

## Icebox

Web app · native mobile · family/multi-user · bank/calendar sync · GraphQL · STT
Mini App UX plan (detail): [docs/miniapp/UX_UI_PLAN.md](../miniapp/UX_UI_PLAN.md)

---

## Decision Gates

| Gate | Criteria | Status |
|------|----------|--------|
| G0 → G1 | Runnable + CI | ✅ |
| G1 → G2 | M1–M4 | ✅ |
| G2 → G3 | 14-day dogfood | 🚧 |
| G3 → G4 | Phase 2 domains | ✅ |
| Stage1 → Stage2 | WIP merged + P0 bugs closed | ⏳ |
