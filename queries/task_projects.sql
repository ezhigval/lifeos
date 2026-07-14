-- name: InsertTaskProject :exec
INSERT INTO task_projects (task_id, project_id)
VALUES ($1, $2);

-- name: DeleteTaskProjects :exec
DELETE FROM task_projects WHERE task_id = $1;

-- name: ListProjectIDsByTask :many
SELECT project_id FROM task_projects WHERE task_id = $1;

-- name: ListTasksByProjectJoin :many
SELECT t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.completed_at, t.deleted_at, t.created_at, t.updated_at, t.duration_minutes, t.tags
FROM tasks t
JOIN task_projects tp ON tp.task_id = t.id
WHERE t.user_id = $1
  AND tp.project_id = $2
  AND t.deleted_at IS NULL
  AND t.status IN ('todo', 'in_progress')
ORDER BY t.due_date ASC NULLS LAST, t.created_at ASC;
