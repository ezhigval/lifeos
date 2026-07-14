-- name: ListOverdueAndTodayTasks :many
SELECT id, user_id, title, description, status, priority, due_date,
       completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status IN ('todo', 'in_progress')
  AND (due_date <= $2 OR priority IN ('urgent', 'high'))
ORDER BY
    CASE priority WHEN 'urgent' THEN 4 WHEN 'high' THEN 3 WHEN 'medium' THEN 2 ELSE 1 END DESC,
    due_date ASC NULLS LAST,
    created_at ASC;

-- name: ListTasksForDay :many
SELECT id, user_id, title, description, status, priority, due_date,
       completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND due_date = $2
  AND status IN ('todo', 'in_progress')
ORDER BY
    CASE priority WHEN 'urgent' THEN 4 WHEN 'high' THEN 3 WHEN 'medium' THEN 2 ELSE 1 END DESC,
    created_at ASC;
