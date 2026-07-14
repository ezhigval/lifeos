-- +goose Up
CREATE TABLE health_weight_logs (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weight_kg  DOUBLE PRECISION NOT NULL,
    logged_at  TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT health_weight_kg_check CHECK (weight_kg > 0 AND weight_kg < 500)
);

CREATE INDEX health_weight_user_logged_idx ON health_weight_logs (user_id, logged_at DESC);

-- +goose Down
DROP TABLE IF EXISTS health_weight_logs;
