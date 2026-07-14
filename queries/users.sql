-- name: GetUserByTelegramID :one
SELECT id, telegram_id, display_name, timezone, created_at, updated_at
FROM users
WHERE telegram_id = $1;

-- name: GetUserByID :one
SELECT id, telegram_id, display_name, timezone, created_at, updated_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, telegram_id, display_name, timezone, created_at, updated_at
FROM users
ORDER BY created_at;
