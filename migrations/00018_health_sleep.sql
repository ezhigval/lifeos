-- +goose Up
CREATE TABLE health_sleep_logs (
    id               UUID PRIMARY KEY,
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    duration_minutes INTEGER NOT NULL,
    logged_at        TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT health_sleep_duration_check CHECK (duration_minutes > 0 AND duration_minutes <= 1440)
);

CREATE INDEX health_sleep_user_logged_idx ON health_sleep_logs (user_id, logged_at DESC);

-- +goose Down
DROP TABLE IF EXISTS health_sleep_logs;
