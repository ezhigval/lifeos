# TASK-003 — Stage 2.1 daily-cycle polish (Habits / Calendar / Settings)

**Agent:** frontend  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Scope:** `web/miniapp/**` + this report / inbox  
**Status:** DONE

---

## Verdict

Habits / Calendar / Settings (+ More hub) are usable for the daily Telegram Mini App cycle: empty/error/loading, create+track, create event, settings saves, nested BackButton. `BrowserRouter` basename `/app` + `freezeInitData` + auth/session unchanged. Build + lint green.

---

## Bugs / polish fixed

| Sev | Area | Issue | Fix |
|-----|------|-------|-----|
| **P1** | Settings | Saving one section invalidated settings and `useEffect` re-hydrated **all** time fields → wiped in-progress edits on other sections | Hydrate once from query; save buttons stay dirty-aware |
| **P1** | Settings | Quiet hours unset (`null`) but local defaults `23:00`/`07:00` → Save disabled forever | Treat missing quiet hours as dirty so first save works |
| **P1** | Habits | Track waited on refetch; all rows disabled while any track pending | Optimistic `today_completed` + streak; disable only the row being tracked; haptic success |
| **P1** | Habits / Calendar / Settings | Create/save failures only haptic — silent dead end | `ruApiError()` + inline RU alerts (incl. sphere 409 «linked projects») |
| **P1** | More | Daily items buried under generic groups; Settings far from cycle | Hub section «Ежедневный цикл» (Привычки → Календарь → Настройки); clearer subtitle |
| **P2** | AppShell | Trailing-slash nested paths could miss BackButton | Normalize path before nested/fallback check (`/more/*` → `/more`) |
| **P2** | Habits / Calendar | Empty list still showed top Create row + EmptyState CTA | Hide header Create when empty; EmptyState owns the CTA |
| **P2** | Calendar | Fixed `12:00` default | Default to next local hour; date subtitle |

API field mapping vs live handlers (`habits`: `id/name/today_completed/streak`; calendar `starts_at` RFC3339; settings `hour/minute` + quiet `start_*`/`end_*`; spheres CRUD) matches `types.ts` — no client remap needed. Defensive `Array.isArray` on list payloads.

Finance overview leftover: comment only; Legend empty copy remains soft («нет расходов» / «разбивка недоступна»), no “API not connected”.

---

## Intact (verified)

- `freezeInitData()` before `BrowserRouter basename="/app"`
- JWT session / 401 re-auth path untouched
- Telegram BackButton on nested `/more/*` → history back or `/more`

---

## Cross-zone

| Ask | Agent | Why |
|-----|-------|-----|
| Confirm user timezone vs browser local when creating calendar events | Backend (optional) | Client sends RFC3339 from browser local wall clock; `ListEventsToday` filters by user TZ. Mismatch TZ → event may not appear “today”. |
| OpenAPI `DELETE /settings/spheres/{id}` documents 204; live handler returns 200 JSON | Backend | Client tolerates both; spec lag only |

No Go written in this task.

---

## Build / lint

```text
cd web/miniapp && npm run build   # OK
cd web/miniapp && npm run lint    # OK (3 pre-existing only-export-components warnings)
```

---

## Commits

(see git SHA after push)
