# Inbox

Architect кладёт задания сюда. Агент читает свой каталог и берёт задачу со статусом `OPEN`.

```
inbox/
  backend/TASK-NNN-….md
  frontend/TASK-NNN-….md
  telegram/TASK-NNN-….md
```

## Task template

```markdown
# TASK-NNN: Title

- **Agent:** backend | frontend | telegram
- **Status:** DRAFT | OPEN | IN_PROGRESS | BLOCKED | DONE
- **Priority:** P0 | P1 | P2
- **Stage:** bugfix | feature | debt

## Goal
…

## In scope
…

## Out of scope
…

## Acceptance
- [ ] …
- [ ] Report filed under docs/agents/reports/<agent>/TASK-NNN.md
```
