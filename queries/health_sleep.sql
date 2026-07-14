-- name: InsertSleepLog :exec
INSERT INTO health_sleep_logs (id, user_id, duration_minutes, logged_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetLatestSleepByUser :one
SELECT id, user_id, duration_minutes, logged_at, created_at
FROM health_sleep_logs
WHERE user_id = $1
ORDER BY logged_at DESC
LIMIT 1;

-- name: ListRecentSleepByUser :many
SELECT id, user_id, duration_minutes, logged_at, created_at
FROM health_sleep_logs
WHERE user_id = $1
ORDER BY logged_at DESC
LIMIT $2;
