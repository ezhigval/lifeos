-- name: InsertHabit :exec
INSERT INTO habits (id, user_id, name, frequency, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: FindHabitByName :one
SELECT id, user_id, name, frequency, created_at
FROM habits
WHERE user_id = $1 AND lower(name) = lower($2);

-- name: GetHabitByID :one
SELECT id, user_id, name, frequency, created_at
FROM habits
WHERE id = $1 AND user_id = $2;

-- name: ListHabits :many
SELECT id, user_id, name, frequency, created_at
FROM habits
WHERE user_id = $1
ORDER BY created_at ASC;
