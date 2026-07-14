# TASK-005 — Stage 3.0 Post-2.2 dogfood free-audit

**Agent:** frontend  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Scope:** `web/miniapp/**` + this report / inbox  
**Status:** DONE

---

## Verdict

Post–Stage 2.2 dogfood: Home / Spheres / Finance / More / auth remain on `BrowserRouter` + `freezeInitData`. Closed remaining **P1** silent dead-ends and wired **Reminders create** to POST `id`. Build + lint green.

---

## Bugs / polish fixed

| Sev | Area | Issue | Fix |
|-----|------|-------|-----|
| **P1** | Reminders | Create ignored new full Reminder body (`id`/`status` from `a1da86c`); list waited on refetch only | Type `createReminder` → `Reminder`; cache-insert on success by `id` |
| **P1** | Finance (Home) | Income/expense failures silent (no `onError`, form wiped before result) | `ruApiError` in sheet; keep amount until close/success |
| **P1** | Spheres | Create sphere/project/task + archive errors haptic-only | RU alerts; insert created entities by returned `id`; EmptyState CTA when empty |
| **P1** | Task detail | Save / complete / archive / delete haptic-only | Banner via `ruApiError` |
| **P1** | Home tasks | Complete failure haptic-only; lists not array-safe | Action alert; `Array.isArray` on priorities/today |
| **P2** | Habits track | Optimistic rollback without RU copy | `actionError` on track failure |
| **P2** | Spheres lists | Non-array payloads possible | Defensive `Array.isArray` on spheres/projects/tasks |

---

## Intact (verified)

- `freezeInitData()` before `BrowserRouter basename="/app"`
- JWT session / 401 re-auth path untouched
- More-domain screens from TASK-004 (Notes/Health/Career/Debts/Analytics) left as-is

---

## Cross-zone

| Ask | Agent | Why |
|-----|-------|-----|
| — | — | Reminders create id delivered by backend `a1da86c`; consumed here. No new blocks. |

No Go written in this task.

---

## Build / lint

```text
cd web/miniapp && npm run build   # OK
cd web/miniapp && npm run lint    # OK (3 pre-existing only-export-components warnings)
```

---

## Commits

- *(filled after commit)*
