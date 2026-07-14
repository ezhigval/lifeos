-- name: InsertCalendarEvent :exec
INSERT INTO calendar_events (id, user_id, title, starts_at, ends_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListCalendarEventsBetween :many
SELECT id, user_id, title, starts_at, ends_at, created_at
FROM calendar_events
WHERE user_id = $1 AND starts_at >= $2 AND starts_at < $3
ORDER BY starts_at ASC;
