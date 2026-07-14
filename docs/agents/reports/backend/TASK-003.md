# Report TASK-003

- **Agent:** backend
- **Status:** DONE
- **Branch / commits:** `cursor/docs-sync-orchestration-fe85` (commit SHA after push)

## Summary

Verified Habits / Calendar / Settings (/spheres) HTTP JSON against Mini App `types.ts` / `client.ts`. Core shapes were already aligned; hardened missing-habit **404**, OpenAPI response schemas, and contract tests.

## Contract (verified OK)

| Endpoint | Mini App expectation | Backend |
|---|---|---|
| `GET /habits/today` | `{habits:[{id,name,today_completed,streak}]}` | match |
| `POST /habits` | `{id,name,frequency}` 201 | match |
| `POST /habits/{id}/track` | `{name,streak}` ; missing → 404 | match (404 hardened) |
| `GET /calendar/today` | `{events:[{id,title,starts_at}]}` | match (RFC3339) |
| `POST /calendar/events` | body `title`+`starts_at` → event 201 | match |
| `GET /settings` | morning/evening/weekly/monthly + quiet hours + language | match |
| spheres CRUD | `id,name,sort_order,created_at`; DELETE **200** body | match |

## Changes

- `internal/transport/http/api/habits.go` — track → 404 on `ErrNotFound`; nil-safe List/Track
- `internal/transport/http/api/{calendar,settings}.go` — nil-safe ListCalendar / GetSettings
- `docs/api/openapi.yaml` — HabitDay / Habit / CalendarEvent / TimeOfDay / UserSettings; DELETE sphere 200+409
- `internal/transport/http/api/api_test.go` — full Habits/Calendar/Settings/Spheres contract coverage + fakes

## Tests run

```
go test ./internal/settings/... ./internal/habits/... ./internal/calendar/... ./internal/spheres/... ./internal/transport/http/api/...
```

All green.

## Cross-zone asks

None. Frontend can rely on these routes as documented.
