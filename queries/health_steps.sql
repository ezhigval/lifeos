-- name: InsertStepLog :exec
INSERT INTO health_step_logs (id, user_id, steps, logged_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetLatestStepsByUser :one
SELECT id, user_id, steps, logged_at, created_at
FROM health_step_logs
WHERE user_id = $1
ORDER BY logged_at DESC
LIMIT 1;

-- name: ListRecentStepsByUser :many
SELECT id, user_id, steps, logged_at, created_at
FROM health_step_logs
WHERE user_id = $1
ORDER BY logged_at DESC
LIMIT $2;
