-- name: InsertDebt :exec
INSERT INTO debts (id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListOpenDebts :many
SELECT id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at
FROM debts
WHERE user_id = $1 AND status = 'open'
LIMIT 1;

-- name: GetDebtByID :one
SELECT id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at
FROM debts
WHERE id = $1 AND user_id = $2;

-- name: FindOpenDebtByCreditor :one
SELECT id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at
FROM debts
WHERE user_id = $1 AND status = 'open' AND creditor ILIKE '%' || $2 || '%'
ORDER BY
    CASE WHEN lower(creditor) = lower($2) THEN 0 ELSE 1 END,
    created_at ASC
LIMIT 1;

-- name: UpdateDebt :exec
UPDATE debts
SET paid_cents = $3, status = $4
WHERE id = $1 AND user_id = $2;
