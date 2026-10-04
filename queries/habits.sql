-- name: InsertHabit :exec
INSERT INTO habits (id, user_id, name, frequency, start_date, end_date, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: FindHabitByName :one
SELECT id, user_id, name, frequency, start_date, end_date, created_at
FROM habits
WHERE user_id = $1 AND lower(name) = lower($2);

-- name: GetHabitByID :one
SELECT id, user_id, name, frequency, start_date, end_date, created_at
FROM habits
WHERE id = $1 AND user_id = $2;

-- name: ListHabits :many
SELECT id, user_id, name, frequency, start_date, end_date, created_at
FROM habits
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: UpdateHabit :execrows
UPDATE habits
SET name = $3,
    frequency = $4,
    start_date = $5,
    end_date = $6
WHERE id = $1 AND user_id = $2;

-- name: DeleteHabit :execrows
DELETE FROM habits
WHERE id = $1 AND user_id = $2;
