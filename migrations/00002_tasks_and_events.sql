-- +goose Up
CREATE TABLE tasks (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title        VARCHAR(500) NOT NULL,
    description  TEXT,
    status       VARCHAR(32) NOT NULL DEFAULT 'todo',
    priority     VARCHAR(16) NOT NULL DEFAULT 'medium',
    due_date     DATE,
    goal_id      UUID,
    completed_at TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tasks_status_check CHECK (status IN ('todo', 'in_progress', 'done', 'cancelled')),
    CONSTRAINT tasks_priority_check CHECK (priority IN ('low', 'medium', 'high', 'urgent'))
);

CREATE INDEX tasks_user_due_date_idx ON tasks (user_id, due_date) WHERE deleted_at IS NULL;
CREATE INDEX tasks_user_status_idx ON tasks (user_id, status) WHERE deleted_at IS NULL;

CREATE TABLE domain_events (
    id             UUID PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id   UUID NOT NULL,
    event_type     VARCHAR(64) NOT NULL,
    payload        JSONB NOT NULL DEFAULT '{}',
    source         VARCHAR(32) NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT domain_events_source_check CHECK (source IN ('telegram', 'scheduler', 'cli', 'http'))
);

CREATE INDEX domain_events_aggregate_idx ON domain_events (aggregate_type, aggregate_id);

-- +goose Down
DROP TABLE IF EXISTS domain_events;
DROP TABLE IF EXISTS tasks;
