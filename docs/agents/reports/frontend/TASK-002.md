# TASK-002 — Stage 2 finance overview wire + polish

**Agent:** frontend  
**Date:** 2026-07-14  
**Branch:** `cursor/docs-sync-orchestration-fe85`  
**Scope:** `web/miniapp/**` + this report / inbox  
**Status:** DONE

---

## Verdict

Client prefers `GET /api/v1/finance/overview?period=YYYY-MM` when 200; cash-flow fallback only on 404/501. Legend/Ring render categories when present; empty copy no longer claims “API not connected”. `BrowserRouter` basename `/app` + `freezeInitData` unchanged. Build + lint green.

---

## Finance wire

| Item | Change |
|------|--------|
| `api.financeOverview` | Normalizes overview payload (`categories` always array); still falls back to cash-flow only for 404/501 |
| `FinanceCard` | Uses enriched overview from Home query as-is (no double enrich); passes `expenseCents` to legend |
| `FinanceLegend` | Shows category bars when `categories.length > 0`; empty → «Нет расходов…» / «Разбивка по категориям пока недоступна» |

Contract matched to Backend overview JSON (`period_label`, `*_cents`, `categories[].name|amount_cents|percent|color_hint`).

---

## Polish (P1 daily-use)

| Area | Fix |
|------|-----|
| Settings → spheres empty | EmptyState CTA «Создать» opens create sheet |
| Project detail empty tasks | EmptyState + «Создать» instead of plain hint text |

Habits / Calendar / TaskDetail / More / BackButton / auth freeze: re-checked after TASK-001 — no additional clear bugs.

---

## Out of scope (untouched)

- Go / OpenAPI / Telegram package (Backend had concurrent uncommitted overview work — not included in this commit)

---

## Build / lint

```text
cd web/miniapp && npm run build   # OK
cd web/miniapp && npm run lint    # OK (3 pre-existing only-export-components warnings)
```

---

## Ask

None if Backend TASK-002 is merged/deployed. If overview still 404/501 in a given environment, UI keeps cash-flow totals without category ring (graceful).
