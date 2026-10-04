-- name: InsertTaskSphere :exec
INSERT INTO task_spheres (task_id, sphere_id)
VALUES ($1, $2);

-- name: DeleteTaskSpheres :exec
DELETE FROM task_spheres WHERE task_id = $1;

-- name: ListSphereIDsByTask :many
SELECT sphere_id FROM task_spheres WHERE task_id = $1;
