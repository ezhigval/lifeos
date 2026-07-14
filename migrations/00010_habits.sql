-- +goose Up
CREATE TABLE habits (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(256) NOT NULL,
    frequency  VARCHAR(16) NOT NULL DEFAULT 'daily',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT habits_frequency_check CHECK (frequency IN ('daily')),
    UNIQUE (user_id, name)
);

CREATE TABLE habit_logs (
    id         UUID PRIMARY KEY,
    habit_id   UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    log_date   DATE NOT NULL,
    completed  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (habit_id, log_date)
);

CREATE INDEX habit_logs_habit_date_idx ON habit_logs (habit_id, log_date DESC);

-- +goose Down
DROP TABLE IF EXISTS habit_logs;
DROP TABLE IF EXISTS habits;
