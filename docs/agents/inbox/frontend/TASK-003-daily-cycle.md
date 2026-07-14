# TASK-003: Stage 2.1 — Mini App daily-cycle polish

- **Agent:** frontend
- **Status:** OPEN
- **Priority:** P1
- **Stage:** miniapp-depth (Stage 2.1)

## Goal

Сделать Habits / Calendar / Settings (+ Home↔More) пригодными для ежедневного юза в Telegram Mini App. Закрыть P0/P1 UX-баги, пустые экраны, маппинг полей API, навигацию.

## In scope

- `web/miniapp/**` only
- Habits: list today, create, track, empty/error, haptic on track
- Calendar: today events, create event, empty/error
- Settings: morning/evening/quiet hours, spheres CRUD (already started) — fix remaining bugs
- More hub clarity; BackButton on nested More routes
- Keep BrowserRouter `/app` + `freezeInitData` + session/auth intact
- Soften/remove leftover finance overview fallback messaging if still present when categories work

## Out of scope

- Go / OpenAPI
- Telegram bot package
- Full parity with bot NL capture

## Acceptance

- [ ] Habits / Calendar / Settings usable for daily flow without dead ends
- [ ] `npm run build` (+ lint) green
- [ ] Report: `docs/agents/reports/frontend/TASK-003.md`
- [ ] Commit + push `cursor/docs-sync-orchestration-fe85`
