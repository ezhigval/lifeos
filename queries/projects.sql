-- name: InsertProject :exec
INSERT INTO projects (id, user_id, name, description, status, outcome, target_value, current_value, unit, target_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: FindProjectByName :one
SELECT id, user_id, name, description, status, outcome, target_value, current_value, unit, target_date, created_at, updated_at
FROM projects
WHERE user_id = $1 AND lower(name) = lower($2) AND status = 'active';

-- name: ListActiveProjects :many
SELECT id, user_id, name, description, status, outcome, target_value, current_value, unit, target_date, created_at, updated_at
FROM projects
WHERE user_id = $1 AND status = 'active'
ORDER BY created_at ASC;

-- name: ProjectExists :one
SELECT EXISTS (
    SELECT 1 FROM projects WHERE id = $1 AND user_id = $2 AND status = 'active'
) AS exists;

-- name: ProjectsExist :one
SELECT count(*)::int = sqlc.arg(expected)::int AS all_exist
FROM projects
WHERE user_id = sqlc.arg(user_id)
  AND status = 'active'
  AND id = ANY(sqlc.arg(project_ids)::uuid[]);

-- name: GetProjectByID :one
SELECT id, user_id, name, description, status, outcome, target_value, current_value, unit, target_date, created_at, updated_at
FROM projects
WHERE id = $1 AND user_id = $2;

-- name: UpdateProjectStatus :exec
UPDATE projects
SET status = $3, updated_at = now()
WHERE id = $1 AND user_id = $2;
