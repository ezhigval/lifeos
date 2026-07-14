-- name: InsertContact :exec
INSERT INTO career_contacts (id, user_id, name, company, role, notes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListRecentContactsByUser :many
SELECT id, user_id, name, company, role, notes, created_at, updated_at
FROM career_contacts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: SearchContactsByUser :many
SELECT id, user_id, name, company, role, notes, created_at, updated_at
FROM career_contacts
WHERE user_id = sqlc.arg(user_id)
  AND (
    name ILIKE '%' || sqlc.arg(query) || '%'
    OR company ILIKE '%' || sqlc.arg(query) || '%'
    OR role ILIKE '%' || sqlc.arg(query) || '%'
  )
ORDER BY created_at DESC
LIMIT sqlc.arg(result_limit);

-- name: DeleteContactByUser :one
DELETE FROM career_contacts
WHERE id = sqlc.arg(contact_id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, company, role, notes, created_at, updated_at;
