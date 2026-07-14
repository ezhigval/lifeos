-- +goose Up
CREATE TABLE day_availability (
    id               UUID PRIMARY KEY,
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day              DATE NOT NULL,
    available_until  TIME,
    note             TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, day)
);

-- +goose Down
DROP TABLE IF EXISTS day_availability;
