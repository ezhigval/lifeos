-- name: InsertNote :exec
INSERT INTO notes (id, user_id, body, tags, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetNoteByID :one
SELECT id, user_id, body, tags, created_at, updated_at
FROM notes
WHERE id = $1 AND user_id = $2;

-- name: UpdateNoteBody :one
UPDATE notes
SET body = $3, updated_at = $4
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, body, tags, created_at, updated_at;

-- name: ListRecentNotesByUser :many
SELECT id, user_id, body, tags, created_at, updated_at
FROM notes
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListNotesByTag :many
SELECT id, user_id, body, tags, created_at, updated_at
FROM notes
WHERE user_id = sqlc.arg(user_id) AND sqlc.arg(tag)::text = ANY(tags)
ORDER BY created_at DESC
LIMIT sqlc.arg(result_limit);

-- name: SearchNotesByUser :many
SELECT id, user_id, body, tags, created_at, updated_at
FROM notes
WHERE user_id = sqlc.arg(user_id) AND body ILIKE '%' || sqlc.arg(query) || '%'
ORDER BY created_at DESC
LIMIT sqlc.arg(result_limit);

-- name: DeleteNoteByUser :one
DELETE FROM notes
WHERE id = sqlc.arg(note_id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, body, tags, created_at, updated_at;
