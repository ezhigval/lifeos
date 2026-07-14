# TASK-002: Stage 2 Mini App polish

- **Agent:** frontend
- **Status:** DONE
- **Priority:** P1
- **Stage:** miniapp-functionality (Stage 2)

## Goal

Wire Mini App to finance overview when Backend ships it; polish Stage 2 UX per `docs/miniapp/UX_UI_PLAN.md`. Prefer real categories over cash-flow fallback.

## In scope

- `web/miniapp/**` only
- Remove/soften fallback once `/finance/overview` returns 200 with categories
- Finance legend/ring when categories present
- Fix any remaining P1 polish on Habits/Calendar/Settings/TaskDetail/More that blocks daily use
- Keep BrowserRouter + initData freeze intact

## Out of scope

- Go / OpenAPI
- Inventing APIs — ask Backend in report if still missing

## Depends on

Backend TASK-002 (can start polish in parallel; pull latest before final wire).

## Acceptance

- [x] FinanceOverview used when endpoint works; categories visible
- [x] `npm run build` (+ lint) green
- [x] Report: `docs/agents/reports/frontend/TASK-002.md`
- [x] Commit + push branch
