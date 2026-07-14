-- name: InsertProcessedUpdate :exec
INSERT INTO processed_updates (update_id) VALUES ($1);

-- name: ProcessedUpdateExists :one
SELECT EXISTS (
    SELECT 1 FROM processed_updates WHERE update_id = $1
) AS exists;
