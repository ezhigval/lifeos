-- name: CountCompletedTasksBetween :one
SELECT COUNT(*)::bigint AS total
FROM tasks
WHERE user_id = $1
  AND status = 'done'
  AND completed_at >= $2
  AND completed_at < $3
  AND deleted_at IS NULL;

-- name: CountOpenTasks :one
SELECT COUNT(*)::bigint AS total
FROM tasks
WHERE user_id = $1
  AND status IN ('todo', 'in_progress')
  AND deleted_at IS NULL;

-- name: CountHabitCompletionsBetween :one
SELECT COUNT(*)::bigint AS total
FROM habit_logs hl
JOIN habits h ON h.id = hl.habit_id
WHERE h.user_id = $1
  AND hl.completed = true
  AND hl.log_date >= $2
  AND hl.log_date < $3;
