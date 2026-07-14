-- name: InsertSphere :exec
INSERT INTO life_spheres (id, user_id, name, sort_order, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListSpheresByUser :many
SELECT id, user_id, name, sort_order, created_at, updated_at
FROM life_spheres
WHERE user_id = $1
ORDER BY sort_order ASC, created_at ASC;

-- name: GetSphereByUser :one
SELECT id, user_id, name, sort_order, created_at, updated_at
FROM life_spheres
WHERE id = $1 AND user_id = $2;

-- name: FindSphereByName :one
SELECT id, user_id, name, sort_order, created_at, updated_at
FROM life_spheres
WHERE user_id = sqlc.arg(user_id) AND lower(name) = lower(sqlc.arg(name));

-- name: CountSpheresByUser :one
SELECT count(*)::int FROM life_spheres WHERE user_id = $1;

-- name: UpdateSphere :exec
UPDATE life_spheres
SET name = $3, sort_order = $4, updated_at = $5
WHERE id = $1 AND user_id = $2;

-- name: DeleteSphereByUser :one
DELETE FROM life_spheres
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, sort_order, created_at, updated_at;
