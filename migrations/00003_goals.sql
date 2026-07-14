-- +goose Up
CREATE TABLE goals (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_goal_id  UUID REFERENCES goals(id) ON DELETE SET NULL,
    title           VARCHAR(500) NOT NULL,
    description     TEXT,
    level           VARCHAR(32) NOT NULL DEFAULT 'month',
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    target_value    NUMERIC,
    current_value   NUMERIC NOT NULL DEFAULT 0,
    unit            VARCHAR(32),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT goals_level_check CHECK (level IN ('life_vision', 'three_year', 'annual', 'quarter', 'month')),
    CONSTRAINT goals_status_check CHECK (status IN ('active', 'achieved', 'abandoned'))
);

CREATE INDEX goals_user_status_idx ON goals (user_id, status);

ALTER TABLE tasks
    ADD CONSTRAINT tasks_goal_id_fkey FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_goal_id_fkey;
DROP TABLE IF EXISTS goals;
