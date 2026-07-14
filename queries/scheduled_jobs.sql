-- name: HasPendingJob :one
SELECT EXISTS (
    SELECT 1 FROM scheduled_jobs
    WHERE user_id = $1 AND job_type = $2 AND status = 'pending'
) AS exists;

-- name: InsertScheduledJob :one
INSERT INTO scheduled_jobs (id, user_id, job_type, payload, run_at, channel)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: ClaimDueJobs :many
WITH picked AS (
    SELECT sj.id
    FROM scheduled_jobs sj
    WHERE sj.status = 'pending' AND sj.run_at <= $1
    ORDER BY sj.run_at
    LIMIT $2
    FOR UPDATE SKIP LOCKED
)
UPDATE scheduled_jobs AS j
SET status = 'processing', updated_at = now()
FROM picked
WHERE j.id = picked.id
RETURNING j.id, j.user_id, j.job_type, j.payload, j.run_at, j.status, j.channel;

-- name: MarkJobDone :exec
UPDATE scheduled_jobs
SET status = 'done', updated_at = now()
WHERE id = $1;

-- name: MarkJobPending :exec
UPDATE scheduled_jobs
SET status = 'pending', updated_at = now(), retry_count = retry_count + 1
WHERE id = $1;

-- name: RescheduleJob :exec
UPDATE scheduled_jobs
SET run_at = $2, status = 'pending', updated_at = now()
WHERE id = $1;

-- name: ReschedulePendingJobsByType :exec
UPDATE scheduled_jobs
SET run_at = $3, status = 'pending', updated_at = now()
WHERE user_id = $1 AND job_type = $2 AND status = 'pending';

-- name: ListPendingRemindersByUser :many
SELECT id, user_id, job_type, payload, run_at, status, channel, created_at
FROM scheduled_jobs
WHERE user_id = $1 AND job_type = 'reminder' AND status = 'pending'
ORDER BY run_at ASC;

-- name: CancelPendingReminder :one
UPDATE scheduled_jobs
SET status = 'cancelled', updated_at = now()
WHERE id = $1 AND user_id = $2 AND job_type = 'reminder' AND status = 'pending'
RETURNING id, user_id, job_type, payload, run_at, status, channel, created_at;
