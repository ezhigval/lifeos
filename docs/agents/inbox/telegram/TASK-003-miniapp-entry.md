# TASK-003: Stage 2.1 — Telegram Mini App entry sanity

- **Agent:** telegram
- **Status:** DONE
- **Priority:** P2
- **Stage:** miniapp-depth (Stage 2.1)

## Goal

Убедиться, что вход в Mini App из бота стабилен после URL-tracking fix: `/start`, reply keyboard web_app, menu button. Только TG-зона.

## In scope

- `internal/transport/telegram/**` (+ notifier if needed)
- `/start` clears stale keyboard flags and attaches current `LIFEOS_MINIAPP_URL`
- No wipe of `reply_kb_miniapp_url` via triage/basePayload
- Docs note in report if ops steps required (`/start` after tunnel rotate)

## Out of scope

- Mini App React
- Finance/habits domain logic

## Acceptance

- [x] Tests green (`ReplyKeyboard`, start/clear if touched)
- [x] Report: `docs/agents/reports/telegram/TASK-003.md`
- [x] Commit only if code/docs change; else report «verified OK»
- [x] Push branch if committed
