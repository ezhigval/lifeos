# Report TASK-004

- **Agent:** backend
- **Status:** DONE
- **Branch / commits:** `cursor/docs-sync-orchestration-fe85` @ `a2f1c09` (`a2f1c099ed17808c50d333a2353e1e4970655498`)

## Summary

Aligned Notes / Health / Career / Reminders / Debts / Analytics HTTP JSON with Mini App `types.ts` / `client.ts`. Core shapes were already correct; hardened 404/400 semantics, analytics `title`/`percent` JSON + empty `projects` array, OpenAPI schemas, and contract tests.

## Contract (verified OK)

| Endpoint | Mini App expectation | Backend |
|---|---|---|
| `GET/POST/DELETE /notes` | `{notes:[{id,body,tags,created_at}]}`; create `{body,tags}`; delete 200 body; missing → 404 | match |
| `GET …/health/*/latest` | WeightLog / StepLog / SleepLog; empty → **404** | match |
| `POST /health/{weight,steps,sleep}` | record body → 201 + log | match (`duration_hours` OK) |
| career contacts/skills CRUD | Contact/Skill shapes; delete 200; missing → 404 | match |
| reminders list/create/delete | `{reminders:[{id,message,fire_at,status}]}`; create ack; cancel 404 | match |
| finance debts list/create/pay | Debt fields; pay → Debt; missing → 404; bad amt → 400 | match |
| `GET /analytics/summary` | percent fields **ints 0–100**; `projects[].title/percent` | match (json tags + empty `[]`) |

## Changes

- `internal/query/analytics.go` — `ProjectKPI` json tags `title` / `percent`
- `internal/transport/http/api/router.go` — analytics nil-safe; coerce nil projects → `[]`; Reminder/Analytics Deps interfaces
- `internal/transport/http/api/{notes,health,career,reminders,finance}.go` — nil-safe handlers; 404/400 preserved
- `docs/api/openapi.yaml` — Note, Debt, Reminder, Contact, Skill, Weight/Step/SleepLog, AnalyticsSummary schemas + status responses
- `internal/transport/http/api/api_test.go` — debt pay, career, health, reminders, analytics contract coverage + fakes

## Tests run

```
go test ./internal/transport/http/api/ ./internal/query/ ./internal/finance/... \
  ./internal/career/... ./internal/health/... ./internal/knowledge/... ./internal/notifications/...
```

All green.

## Cross-zone asks

None material. Frontend already accepts PascalCase analytics project fields as fallback (`Title \|\| title`); lowercase is now canonical.

**Note:** `POST /reminders` returns `{message, fire_at}` ack (not full Reminder); Mini App client does not type the create response.
