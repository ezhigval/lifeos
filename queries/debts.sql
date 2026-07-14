-- name: InsertDebt :exec
INSERT INTO debts (
    id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at,
    installment_cents, installment_interval, next_payment_date
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: ListOpenDebts :many
SELECT id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at,
       installment_cents, installment_interval, next_payment_date
FROM debts
WHERE user_id = $1 AND status = 'open'
ORDER BY created_at ASC;

-- name: GetDebtByID :one
SELECT id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at,
       installment_cents, installment_interval, next_payment_date
FROM debts
WHERE id = $1 AND user_id = $2;

-- name: FindOpenDebtByCreditor :one
SELECT id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at,
       installment_cents, installment_interval, next_payment_date
FROM debts
WHERE user_id = $1 AND status = 'open' AND creditor ILIKE '%' || $2 || '%'
ORDER BY
    CASE WHEN lower(creditor) = lower($2) THEN 0 ELSE 1 END,
    created_at ASC
LIMIT 1;

-- name: UpdateDebt :exec
UPDATE debts
SET paid_cents = $3,
    status = $4,
    installment_cents = $5,
    installment_interval = $6,
    next_payment_date = $7
WHERE id = $1 AND user_id = $2;

-- name: ListPlannedDebtPayments :many
SELECT id, user_id, creditor, amount_cents, paid_cents, due_date, status, created_at,
       installment_cents, installment_interval, next_payment_date
FROM debts
WHERE user_id = $1
  AND status = 'open'
  AND installment_cents > 0
  AND installment_interval <> 'none'
  AND next_payment_date IS NOT NULL
  AND next_payment_date <= sqlc.arg(until_date)
ORDER BY next_payment_date ASC;
