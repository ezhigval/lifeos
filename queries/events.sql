-- name: InsertDomainEvent :exec
INSERT INTO domain_events (
    id, user_id, aggregate_type, aggregate_id, event_type, payload, source, occurred_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);
