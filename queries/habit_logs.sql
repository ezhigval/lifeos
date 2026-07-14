-- name: UpsertHabitLog :exec
INSERT INTO habit_logs (id, habit_id, log_date, completed, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (habit_id, log_date) DO UPDATE
SET completed = EXCLUDED.completed;

-- name: ListHabitLogsSince :many
SELECT id, habit_id, log_date, completed, created_at
FROM habit_logs
WHERE habit_id = $1 AND log_date >= $2
ORDER BY log_date DESC;

-- name: ListHabitsWithTodayLog :many
SELECT
    h.id,
    h.user_id,
    h.name,
    h.frequency,
    h.created_at,
    hl.completed AS today_completed
FROM habits h
LEFT JOIN habit_logs hl ON hl.habit_id = h.id AND hl.log_date = $2
WHERE h.user_id = $1
ORDER BY h.created_at ASC;
