-- +goose Up
ALTER TABLE tasks
    ADD COLUMN duration_minutes INT,
    ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

ALTER TABLE tasks
    ADD CONSTRAINT tasks_duration_minutes_check CHECK (duration_minutes IS NULL OR duration_minutes > 0);

CREATE INDEX tasks_tags_gin_idx ON tasks USING GIN (tags);

-- +goose Down
DROP INDEX IF EXISTS tasks_tags_gin_idx;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_duration_minutes_check;
ALTER TABLE tasks DROP COLUMN IF EXISTS tags;
ALTER TABLE tasks DROP COLUMN IF EXISTS duration_minutes;
