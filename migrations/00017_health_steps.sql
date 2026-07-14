-- +goose Up
CREATE TABLE health_step_logs (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    steps      INTEGER NOT NULL,
    logged_at  TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT health_steps_check CHECK (steps > 0 AND steps <= 200000)
);

CREATE INDEX health_steps_user_logged_idx ON health_step_logs (user_id, logged_at DESC);

-- +goose Down
DROP TABLE IF EXISTS health_step_logs;
