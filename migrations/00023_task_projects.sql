-- +goose Up
CREATE TABLE task_projects (
    task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, project_id)
);

CREATE INDEX task_projects_project_idx ON task_projects (project_id);

INSERT INTO task_projects (task_id, project_id)
SELECT id, project_id FROM tasks WHERE project_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS task_projects;
