-- +goose Up
ALTER TABLE projects
    ADD COLUMN outcome       TEXT,
    ADD COLUMN target_value  NUMERIC,
    ADD COLUMN current_value NUMERIC NOT NULL DEFAULT 0,
    ADD COLUMN unit          VARCHAR(32),
    ADD COLUMN target_date   DATE,
    ADD COLUMN updated_at    TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TABLE project_spheres (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sphere_id  UUID NOT NULL REFERENCES life_spheres(id) ON DELETE RESTRICT,
    PRIMARY KEY (project_id, sphere_id)
);

CREATE INDEX project_spheres_sphere_idx ON project_spheres (sphere_id);

INSERT INTO projects (id, user_id, name, description, status, target_value, current_value, unit, created_at, updated_at)
SELECT g.id, g.user_id, g.title, g.description,
    CASE g.status WHEN 'achieved' THEN 'completed' WHEN 'abandoned' THEN 'archived' ELSE 'active' END,
    g.target_value, g.current_value, g.unit, g.created_at, g.updated_at
FROM goals g
WHERE g.id NOT IN (SELECT id FROM projects);

INSERT INTO project_spheres (project_id, sphere_id)
SELECT g.id, ls.id
FROM goals g
JOIN life_spheres ls ON ls.user_id = g.user_id AND ls.name = 'Карьера'
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS project_spheres;
ALTER TABLE projects
    DROP COLUMN IF EXISTS outcome,
    DROP COLUMN IF EXISTS target_value,
    DROP COLUMN IF EXISTS current_value,
    DROP COLUMN IF EXISTS unit,
    DROP COLUMN IF EXISTS target_date,
    DROP COLUMN IF EXISTS updated_at;
