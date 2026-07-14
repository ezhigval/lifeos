-- name: InsertWeightLog :exec
INSERT INTO health_weight_logs (id, user_id, weight_kg, logged_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetLatestWeightByUser :one
SELECT id, user_id, weight_kg, logged_at, created_at
FROM health_weight_logs
WHERE user_id = $1
ORDER BY logged_at DESC
LIMIT 1;

-- name: ListRecentWeightsByUser :many
SELECT id, user_id, weight_kg, logged_at, created_at
FROM health_weight_logs
WHERE user_id = $1
ORDER BY logged_at DESC
LIMIT $2;
