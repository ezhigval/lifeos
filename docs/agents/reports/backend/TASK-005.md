# Report TASK-005

- **Agent:** backend
- **Status:** DONE
- **Branch / commits:** `cursor/docs-sync-orchestration-fe85` @ `5055778` (`5055778d55137c848d0a6a0ed96fc7ab8cfc78cb`)

## Summary

Free-audit of HTTP/API/domain after Stage 2.2 + reminder create id change. Fixed **P1 timezone month/week bounds** (analytics, cash-flow, reviews, finance month totals) that truncated the last calendar day for non-UTC zones (e.g. Europe/Moscow). Hardened reminder create→cancel id contract tests. Fixed flaky list-today timezone test past 21:00 UTC.

## Bugs found / fixed

| Sev | Area | Bug | Fix |
|-----|------|-----|-----|
| **P1** | Analytics / cash-flow / reviews | `MonthBounds` / `PreviousMonthBounds` added months on the UTC instant after converting local midnight → end was **1 day early** for positive offsets (Moscow July ended Jul 30 21:00Z instead of Jul 31 21:00Z) | Compute end in local TZ then `.UTC()`; `MonthBounds` delegates to `MonthBoundsForPeriod` |
| **P1** | Finance month totals | `MonthStartInTimezone` built `time.Date(..., time.UTC)` (wall clock as UTC) + `AddDate` → wrong `[start,end)` vs overview | Return local month-start UTC instant; `finance.monthBounds` uses `MonthBounds` |
| **P1** | Task lifecycle | `TestListTasksTodayUsesTimezone` assumed Jul 14 MSK while wall clock after 21:00 UTC is Jul 15 MSK → empty today list in CI evenings | Derive expected due date from Moscow wall calendar |
| **P0** (guard) | Reminders | Create now returns full ReminderDTO (`id`/`status`); contract test only checked list | Assert create body id/status; cancel by **create** id; list id matches |

## Verified OK (no code change)

| Surface | Result |
|---|---|
| `POST /reminders` → ReminderDTO with id | Already fixed in `a1da86c`; tests strengthened |
| Health `*/latest` empty → **404** | Contract green |
| `GET /analytics/summary` ints + `title`/`percent` + `projects:[]` | Contract green |
| `GET /finance/overview?period=YYYY-MM` | Already used correct `MonthBoundsForPeriod` |
| Task PATCH null-clear / 404 / archive / delete | Lifecycle contract green |

## Changes

- `internal/platform/timeutil/{period,timeutil}.go` + `period_test.go`
- `internal/finance/app/period.go`
- `internal/transport/http/api/api_test.go` — reminder create id contract
- `internal/tasks/app/app_test.go` — list-today timezone freeze

## Tests run

```
go test ./internal/platform/timeutil/ ./internal/finance/... \
  ./internal/transport/http/api/ ./internal/tasks/... \
  ./internal/notifications/... ./internal/health/... ./internal/query/...
go build ./cmd/lifeos
```

All green.

## Cross-zone asks

None blocking. Frontend can optimistic-insert from `POST /reminders` id (DTO already live).
