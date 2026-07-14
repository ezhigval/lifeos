# TASK-001 — Frontend bugfix free-audit

**Agent:** frontend  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Scope:** `web/miniapp/**` only  
**Status:** DONE (P0/P1 in-zone closed)

---

## Verdict

Post–miniapp-ux merge: auth/`BrowserRouter`+initData freeze remains intact. Found and fixed **4 P1** client bugs (analytics %, EditTask clear flags, BackButton, unknown routes). Build + lint OK. No Go/OpenAPI changes.

---

## Bugs found / fixed

| Sev | Area | Bug | Fix |
|-----|------|-----|-----|
| **P1** | Analytics | `completion_rate` / `habit_consistency` already 0–100 ints from API; UI did `* 100` → e.g. `8500%` | `formatPercent()` treats 0–100 (defends 0–1 fraction) |
| **P1** | TaskDetail / API client | PATCH wired to `EditTask`; client sent `due_date: null` / `description: null` which are **no-ops** — clearing due date or description silently failed | Send `clear_due_date` / `clear_description` when empty; align `updateTask` body type |
| **P1** | BackButton | `window.history.length > 1` unreliable in TG WebView; `/tasks/:id` fell back to `/spheres` | Use RR `history.state.idx`; task fallback → `/` |
| **P1** | Routing | Unknown nested paths under `/app` rendered blank shell | `path="*"` → `<Navigate to="/" replace />` |

### Hardened (non-breaking)

- Health sleep display: safe `duration_hours` formatting if field missing.

---

## Verified OK (no fix needed)

- **Auth / initData:** `freezeInitData()` before `BrowserRouter`; JWT restore without reusing initData; 401 → silent re-auth from frozen initData; HashRouter not used.
- **Nav:** BottomNav Главная / Сферы / Ещё; More nested routes; Habits/Calendar/Settings/TaskDetail wired to existing REST paths.
- **Empty/error:** QueryError / EmptyState present on audited screens.
- **archive** path `/tasks/{id}/archive` exists in Go router (alongside `/cancel`).

---

## Asks (out of scope — not blocked)

| Ask | Agent | Why |
|-----|-------|-----|
| `GET /api/v1/finance/overview?period=` | Backend | Client falls back to cash-flow (no categories). Overview still 404/501 → ring without breakdown. |
| OpenAPI: document `POST /tasks/{id}/archive` + EditTask `clear_*` | Backend | Spec lags live router (`archive` + `clear_due_date` / `clear_description`). |
| Priority items lack task `id` | Backend (optional) | Home list can only deep-link priorities that also appear in `/tasks/today` (title match). |

---

## Build / lint

```text
cd web/miniapp && npm run build   # tsc -b && vite build — OK
cd web/miniapp && npm run lint    # oxlint — OK (3 pre-existing only-export-components warnings)
```

---

## Commits

- `627f64fbcef77419c7901af0edb653260b00e743` — `fix(miniapp): close P1 post-merge UX bugs`
- `da3e6901be2c7cd853e4d4c792ab667ac4796d71` — `docs(frontend): correct TASK-001 commit SHA in report`
