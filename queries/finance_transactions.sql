-- name: InsertFinanceTransaction :exec
INSERT INTO finance_transactions (
    id, user_id, category_id, kind, amount_cents, currency, description, occurred_at, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: SumIncomeBetween :one
SELECT COALESCE(SUM(amount_cents), 0)::bigint AS total
FROM finance_transactions
WHERE user_id = $1
  AND kind = 'income'
  AND occurred_at >= $2
  AND occurred_at < $3;

-- name: SumExpenseBetween :one
SELECT COALESCE(SUM(amount_cents), 0)::bigint AS total
FROM finance_transactions
WHERE user_id = $1
  AND kind = 'expense'
  AND occurred_at >= $2
  AND occurred_at < $3;

-- name: SumExpensesByCategoryBetween :many
SELECT c.name AS name, COALESCE(SUM(t.amount_cents), 0)::bigint AS amount_cents
FROM finance_transactions t
JOIN finance_categories c ON c.id = t.category_id
WHERE t.user_id = $1
  AND t.kind = 'expense'
  AND t.occurred_at >= $2
  AND t.occurred_at < $3
GROUP BY c.name
ORDER BY amount_cents DESC, c.name ASC;
