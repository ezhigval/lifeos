-- +goose Up
CREATE TABLE projects (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(256) NOT NULL,
    description TEXT,
    status      VARCHAR(16) NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT projects_status_check CHECK (status IN ('active', 'archived', 'completed')),
    UNIQUE (user_id, name)
);

ALTER TABLE tasks
    ADD COLUMN project_id UUID REFERENCES projects(id);

CREATE INDEX tasks_user_project_idx ON tasks (user_id, project_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS tasks_user_project_idx;
ALTER TABLE tasks DROP COLUMN IF EXISTS project_id;
DROP TABLE IF EXISTS projects;
