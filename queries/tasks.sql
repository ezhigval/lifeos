-- name: InsertTask :one
INSERT INTO tasks (
    id, user_id, title, description, status, priority, due_date, duration_minutes, tags, kind, address, note_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13
)
RETURNING id, user_id, title, description, status, priority, due_date, completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id;

-- name: GetTaskByID :one
SELECT id, user_id, title, description, status, priority, due_date, completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id
FROM tasks
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: ListTasksByDueDate :many
SELECT id, user_id, title, description, status, priority, due_date, completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id
FROM tasks
WHERE user_id = $1
  AND due_date = $2
  AND deleted_at IS NULL
ORDER BY
    CASE priority
        WHEN 'urgent' THEN 4
        WHEN 'high' THEN 3
        WHEN 'medium' THEN 2
        WHEN 'low' THEN 1
    END DESC,
    created_at ASC;

-- name: ListOpenTasksDueOnOrBefore :many
SELECT id, user_id, title, description, status, priority, due_date, completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status IN ('todo', 'in_progress')
  AND due_date IS NOT NULL
  AND due_date <= $2
ORDER BY due_date ASC, created_at ASC;

-- name: ListOpenTasksDueBetween :many
SELECT id, user_id, title, description, status, priority, due_date, completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status IN ('todo', 'in_progress')
  AND due_date IS NOT NULL
  AND due_date >= sqlc.arg(from_date)
  AND due_date <= sqlc.arg(to_date)
ORDER BY due_date ASC, created_at ASC;

-- name: ListTasksByTag :many
SELECT id, user_id, title, description, status, priority, due_date, completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status IN ('todo', 'in_progress')
  AND sqlc.arg(tag)::text = ANY(tags)
ORDER BY due_date ASC NULLS LAST, created_at ASC;

-- name: FindOpenTaskByTitle :one
SELECT id, user_id, title, description, status, priority, due_date, completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status IN ('todo', 'in_progress')
  AND title ILIKE '%' || $2 || '%'
ORDER BY
    CASE WHEN lower(title) = lower($2) THEN 0 ELSE 1 END,
    due_date ASC NULLS LAST,
    created_at ASC
LIMIT 1;

-- name: UpdateTask :one
UPDATE tasks
SET title = $3,
    description = $4,
    status = $5,
    priority = $6,
    due_date = $7,
    duration_minutes = $8,
    tags = $9,
    kind = $10,
    address = $11,
    note_id = $12,
    completed_at = $13,
    deleted_at = $14,
    updated_at = now()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, user_id, title, description, status, priority, due_date, completed_at, deleted_at, created_at, updated_at, duration_minutes, tags, kind, address, note_id;
