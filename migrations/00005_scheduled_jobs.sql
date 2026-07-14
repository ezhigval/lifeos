-- +goose Up
CREATE TABLE scheduled_jobs (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_type    VARCHAR(64) NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    run_at      TIMESTAMPTZ NOT NULL,
    status      VARCHAR(32) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    channel     VARCHAR(32) NOT NULL DEFAULT 'telegram',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT scheduled_jobs_status_check CHECK (status IN ('pending', 'processing', 'done', 'cancelled')),
    CONSTRAINT scheduled_jobs_job_type_check CHECK (job_type IN ('reminder', 'morning_review', 'evening_review'))
);

CREATE INDEX scheduled_jobs_run_at_status_idx ON scheduled_jobs (run_at, status) WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS scheduled_jobs;
