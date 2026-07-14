# Agents — Orchestration

**Architect** оркестрирует трёх исполнителей. Общение — через эту папку (и Cursor Task agents при запуске).

| Роль | Файл | Зона кода |
|------|------|-----------|
| Architect (BOSS) | [ARCHITECT.md](ARCHITECT.md) | docs sync, roadmap, inbox assignments, review отчётов |
| Backend | [BACKEND.md](BACKEND.md) | `internal/*/domain|app|infra`, `query`, `ai`, `platform`, HTTP API, migrations/sqlc |
| Frontend | [FRONTEND.md](FRONTEND.md) | `web/miniapp/**` только |
| Telegram | [TELEGRAM.md](TELEGRAM.md) | `internal/transport/telegram/**`, TG notifier, Bot UX |

## Workflow

```
Architect пишет задание → docs/agents/inbox/<agent>/TASK-NNN.md
Agent делает работу только в своей зоне
Agent пишет отчёт → docs/agents/reports/<agent>/TASK-NNN.md
Architect ревьюит → CLOSED / REWORK в том же отчёте
```

## Rules

1. Один агент = одна зона. Пересечения (контракты API, callback_data, OpenAPI) согласуются через Architect.
2. Source of truth: код + `docs/architecture/*` + OpenAPI. При расхождении — чинить docs в том же изменении.
3. Stage 1 после планирования: **bugfix**. Каждый агент чинит только свои баги.
4. Не мержить чужие ветки / не трогать чужой tree «заодно».

## Snapshots

- [CURRENT_STATE.md](CURRENT_STATE.md) — что есть в коде сейчас
- [PLANNING_NOTES.md](PLANNING_NOTES.md) — черновик для совместного roadmap / bugfix
- [inbox/](inbox/) · [reports/](reports/)
