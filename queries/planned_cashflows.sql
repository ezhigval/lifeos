-- name: InsertPlannedCashflow :exec
INSERT INTO planned_cashflows (
    id, user_id, kind, title, amount_cents, interval, next_date, debt_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $9
);

-- name: ListPlannedCashflowsByUser :many
SELECT id, user_id, kind, title, amount_cents, interval, next_date, debt_id, created_at, updated_at
FROM planned_cashflows
WHERE user_id = $1
ORDER BY next_date ASC, created_at ASC;

-- name: DeletePlannedCashflowByUser :one
DELETE FROM planned_cashflows
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING id, user_id, kind, title, amount_cents, interval, next_date, debt_id, created_at, updated_at;
