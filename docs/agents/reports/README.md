# Reports

Агент пишет отчёт после задачи (или при BLOCKED).

```
reports/
  backend/TASK-NNN.md
  frontend/TASK-NNN.md
  telegram/TASK-NNN.md
```

## Report template

```markdown
# Report TASK-NNN

- **Agent:** …
- **Status:** DONE | BLOCKED | NEED_REWORK
- **Branch / commits:** …

## Summary
…

## Changes
- paths…

## Tests run
…

## Cross-zone asks
(если нужны API / кнопки / UI от другого агента)

## Architect review
- [ ] ACK / REWORK notes
```
