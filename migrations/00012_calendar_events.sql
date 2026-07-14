-- +goose Up
CREATE TABLE calendar_events (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(500) NOT NULL,
    starts_at  TIMESTAMPTZ NOT NULL,
    ends_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX calendar_events_user_starts_idx ON calendar_events (user_id, starts_at);

-- +goose Down
DROP TABLE IF EXISTS calendar_events;
