-- +goose Up
ALTER TABLE habits
    ADD COLUMN start_date DATE NULL,
    ADD COLUMN end_date   DATE NULL;

CREATE INDEX habits_user_id_idx ON habits (user_id);

-- +goose Down
DROP INDEX IF EXISTS habits_user_id_idx;
ALTER TABLE habits
    DROP COLUMN IF EXISTS start_date,
    DROP COLUMN IF EXISTS end_date;
