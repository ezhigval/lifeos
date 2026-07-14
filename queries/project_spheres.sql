-- name: InsertProjectSphere :exec
INSERT INTO project_spheres (project_id, sphere_id)
VALUES ($1, $2);

-- name: DeleteProjectSpheres :exec
DELETE FROM project_spheres WHERE project_id = $1;

-- name: ListSphereIDsByProject :many
SELECT sphere_id FROM project_spheres WHERE project_id = $1;

-- name: ListProjectsBySphere :many
SELECT p.id, p.user_id, p.name, p.description, p.status, p.outcome, p.target_value, p.current_value, p.unit, p.target_date, p.created_at, p.updated_at
FROM projects p
JOIN project_spheres ps ON ps.project_id = p.id
WHERE p.user_id = sqlc.arg(user_id) AND ps.sphere_id = sqlc.arg(sphere_id) AND p.status = 'active'
ORDER BY p.created_at ASC;

-- name: CountProjectsBySphere :one
SELECT count(*)::int
FROM project_spheres ps
JOIN projects p ON p.id = ps.project_id
WHERE ps.sphere_id = $1 AND p.status = 'active';
