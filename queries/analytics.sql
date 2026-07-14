-- name: CountTasksCreatedBetween :one
SELECT COUNT(*)::bigint AS total
FROM tasks
WHERE user_id = $1
  AND created_at >= $2
  AND created_at < $3
  AND deleted_at IS NULL;

-- name: CountUserHabits :one
SELECT COUNT(*)::bigint AS total
FROM habits
WHERE user_id = $1;
