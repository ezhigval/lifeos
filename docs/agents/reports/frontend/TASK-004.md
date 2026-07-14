# TASK-004 — Stage 2.2 More domains polish (Notes / Health / Career / Reminders / Debts / Analytics)

**Agent:** frontend  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Scope:** `web/miniapp/**` + this report / inbox  
**Status:** DONE

---

## Verdict

Notes, Health, Career, Reminders, Debts, Analytics (+ More hub regroup) are usable for daily Mini App use: empty/error/loading, create/delete with validation + RU alerts, optimistic delete/pay/create where safe, haptics. `BrowserRouter` basename `/app` + `freezeInitData` + auth/session unchanged. Build + lint green.

---

## Bugs / polish fixed

| Sev | Area | Issue | Fix |
|-----|------|-------|-----|
| **P1** | All create sheets | API failures only haptic — silent dead end | `ruApiError()` + inline RU alerts |
| **P1** | Health | 404 «no records» treated as query error; empty vs hard error muddy | Map 404 → `null` data; show metrics + hard-error Retry separately |
| **P1** | Health | Invalid input / save errors silent | Client validation + formError; optimistic latest cache on save |
| **P1** | Reminders | Past `fire_at` / bad datetime → opaque failure | Client future-check + RFC3339 map; RU status labels |
| **P1** | Notes / Career / Reminders / Debts | Empty lists still showed header Create; no EmptyState CTA | Hide Create when empty; EmptyState owns CTA |
| **P1** | Deletes / cancel / pay | Waited on refetch; no rollback | Optimistic remove (notes/contacts/skills/reminders); optimistic pay remaining; per-row busy |
| **P1** | Debts | Overpay / pay errors silent | Client remaining check + `ruApiError` (overpayment / not open) |
| **P2** | Analytics | EN «Completion»; Go `Title`/`Percent` without tags | RU labels; normalize `Title\|title` / `Percent\|percent` |
| **P2** | Lists | Non-array payloads possible | `Array.isArray` defensive on notes/reminders/debts/contacts/skills |
| **P2** | More hub | Reminders buried in «Обзор»; Domains mixed | Groups: **Ежедневный цикл** (+Reminders) · **Запись** (Notes/Health/Debts) · **Обзор** (Career/Analytics) |

API field mapping matches live handlers (`body/created_at`, health `weight_kg`/`steps`/`duration_hours`, debts `remaining_cents`/`due_date` YYYY-MM-DD, reminders `fire_at` RFC3339, career name/company/role/level). No client remaps needed beyond analytics ProjectKPI.

---

## Intact (verified)

- `freezeInitData()` before `BrowserRouter basename="/app"`
- JWT session / 401 re-auth path untouched
- Telegram BackButton on nested `/more/*` unchanged

---

## Cross-zone

| Ask | Agent | Why |
|-----|-------|-----|
| `POST /reminders` returns `{message,fire_at}` without `id`/`status` | Backend | Client cannot optimistic-insert; must invalidate+refetch. Prefer full `reminderJSON` like list/cancel. |
| Analytics `projects[]` marshals Go `ProjectKPI` without `json` tags → `Title`/`Percent` | Backend (optional) | Client tolerates both casings; add tags for snake/lower consistency with rest of API |

No Go written in this task.

---

## Build / lint

```text
cd web/miniapp && npm run build   # OK
cd web/miniapp && npm run lint    # OK (3 pre-existing only-export-components warnings)
```

---

## Commits

- `af8fc3b22341678f16320b1c69aa3316c743c924` — `fix(miniapp): Stage 2.2 Notes/Health/Career/Reminders/Debts`
