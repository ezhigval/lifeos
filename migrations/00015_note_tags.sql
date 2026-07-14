-- +goose Up
ALTER TABLE notes ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

CREATE INDEX notes_tags_gin_idx ON notes USING GIN (tags);

-- +goose Down
DROP INDEX IF EXISTS notes_tags_gin_idx;
ALTER TABLE notes DROP COLUMN IF EXISTS tags;
