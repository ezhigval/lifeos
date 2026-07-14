-- +goose Up
CREATE TABLE notes (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX notes_user_created_idx ON notes (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS notes;
