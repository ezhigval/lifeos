# Report TASK-001

- **Agent:** backend
- **Status:** DONE
- **Branch / commits:** `cursor/docs-sync-orchestration-fe85` @ `c4435b5` (`c4435b51b0996765d30e0de633945395ce2bf22c`)

## Summary

Free-audit of task lifecycle after WIP merges. Closed P0/P1 bugs in backend zone: Mini App PATCH `null` did not clear `description`/`due_date`; `CreateTask` lacked Description vs TaskDTO; mutation use cases/HTTP returned 400 instead of 404 for missing tasks; `ApplyEdit` allowed editing terminal tasks; OpenAPI missing GET/DELETE/archive and description fields.

## Bugs found / fixed

| Sev | Bug | Fix |
|-----|-----|-----|
| P0 | `PATCH /tasks/{id}` used EditTask with `*string`; Mini App sends `due_date:null` / `description:null` to clear, which Go treated as omit → fields never cleared | `nullableString` presence detection; null/empty → ClearDueDate / ClearDescription |
| P1 | `CreateTaskInput` / HTTP create omitted `Description` while `TaskDTO` and DB expose it | Added Description to create UC + HTTP + OpenAPI |
| P1 | Edit/Cancel/Reschedule/Complete GetByID errors not mapped; HTTP always 400 | Map `ErrNotFound` → `ErrTaskNotFound`; HTTP 404 |
| P1 | `ApplyEdit` (UpdateTask path) allowed editing done/cancelled | Reject with `ErrCannotEditTerminal` |
| P1 | OpenAPI missing GET/DELETE/`/archive`, description/clear on PATCH | Synced `docs/api/openapi.yaml` |
| P2 | Archive/Delete/GetTask nil deps could panic | Nil-guard → 501 |

## Changes

- `internal/tasks/domain/task.go` — ApplyEdit terminal guard
- `internal/tasks/app/{create_task,edit_task,cancel_task,reschedule_task,complete_task}.go`
- `internal/transport/http/api/{router.go,tasks_mutate.go,nullable.go}`
- `docs/api/openapi.yaml`
- Tests: domain, app, HTTP lifecycle + nullableString

## Tests run

```
go test ./internal/tasks/... ./internal/transport/http/api/...
go build ./cmd/lifeos
```

All green.

## Cross-zone asks

None blocking. Frontend already sends Mini App PATCH null-clear payload; backend now matches. Optional FE: use `clear_description` / `clear_due_date` flags as well (redundant if null is sent).

## Architect review

- [ ] ACK / REWORK notes
