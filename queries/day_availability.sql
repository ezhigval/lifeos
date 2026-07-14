-- name: UpsertDayAvailability :one
INSERT INTO day_availability (id, user_id, day, available_until, note, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, day) DO UPDATE
SET available_until = EXCLUDED.available_until,
    note = EXCLUDED.note
RETURNING id, user_id, day, available_until, note, created_at;

-- name: GetDayAvailability :one
SELECT id, user_id, day, available_until, note, created_at
FROM day_availability
WHERE user_id = $1 AND day = $2;
