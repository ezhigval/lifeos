-- name: InsertSkill :exec
INSERT INTO career_skills (id, user_id, name, level, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListRecentSkillsByUser :many
SELECT id, user_id, name, level, created_at, updated_at
FROM career_skills
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: SearchSkillsByUser :many
SELECT id, user_id, name, level, created_at, updated_at
FROM career_skills
WHERE user_id = sqlc.arg(user_id)
  AND (
    name ILIKE '%' || sqlc.arg(query) || '%'
    OR level ILIKE '%' || sqlc.arg(query) || '%'
  )
ORDER BY created_at DESC
LIMIT sqlc.arg(result_limit);

-- name: DeleteSkillByUser :one
DELETE FROM career_skills
WHERE id = sqlc.arg(skill_id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, name, level, created_at, updated_at;
