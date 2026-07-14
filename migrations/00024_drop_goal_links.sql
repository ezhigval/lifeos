-- +goose Up
INSERT INTO task_projects (task_id, project_id)
SELECT id, goal_id FROM tasks WHERE goal_id IS NOT NULL
ON CONFLICT DO NOTHING;

ALTER TABLE tasks DROP COLUMN IF EXISTS goal_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS project_id;

DROP TABLE IF EXISTS goals;

-- +goose Down
-- irreversible: goals data lives in projects after 00022
