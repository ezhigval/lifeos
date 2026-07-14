# TASK-008: Collect & fix dogfood P0 (telegram)

- **Agent:** telegram
- **Status:** DRAFT
- **Priority:** P0
- **Stage:** dogfood (G2→G3)

## Goal

Принять P0/P1 из 14-дневного dogfood Valentin ([DOGFOOD.md](../../../roadmap/DOGFOOD.md)) в зоне бота и закрыть блокеры gate.

## In scope

- Баги из `BUG-DOGFOOD-*` / asks с Agent=telegram
- `/start`, reply keyboard, Mini App web_app button / URL rotate
- NL: task / habit / expense; reminders; morning/evening/weekly/monthly reviews
- Presentation (HTML/format) без ломки capture flow

## Out of scope

- Mini App screens / REST handlers (чужие зоны)
- Полный rewrite strangler / intelligence

## Acceptance

- [ ] Architect → Status OPEN когда есть конкретный P0 backlog
- [ ] P0 в зоне fixed или BLOCKED с ask
- [ ] Report `docs/agents/reports/telegram/TASK-008.md`
- [ ] Commit + push
