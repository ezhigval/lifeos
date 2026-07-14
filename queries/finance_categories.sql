-- name: GetFinanceCategoryByName :one
SELECT id, user_id, name, kind, created_at
FROM finance_categories
WHERE user_id = $1 AND name = $2 AND kind = $3;

-- name: InsertFinanceCategory :exec
INSERT INTO finance_categories (id, user_id, name, kind, created_at)
VALUES ($1, $2, $3, $4, $5);
