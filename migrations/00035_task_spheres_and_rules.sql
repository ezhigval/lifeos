-- +goose Up
-- 00035: N:M задачи↔сферы + привязки сфер к доменным сущностям (TASK-011 п.7-8)
-- spheres нет: справочник сфер — life_spheres (00036 чинил уже созданный FK).

CREATE TABLE IF NOT EXISTS task_spheres (
    task_id   UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    sphere_id UUID NOT NULL REFERENCES life_spheres(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (task_id, sphere_id)
);

CREATE INDEX IF NOT EXISTS idx_task_spheres_sphere ON task_spheres(sphere_id);

-- Связь сферы с внешними сущностями: finance_category / habit / workspace.
-- Используется домен-правилами (зарплата→план дохода, проект здоровья→трекер,
-- карьера→воркспейс-мост) и фильтрами главной/календаря.
CREATE TABLE IF NOT EXISTS sphere_domain_links (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sphere_id  UUID NOT NULL REFERENCES life_spheres(id) ON DELETE CASCADE,
    link_type  TEXT NOT NULL CHECK (link_type IN ('finance_category','habit','workspace')),
    ref_id     UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (sphere_id, link_type, ref_id)
);

CREATE INDEX IF NOT EXISTS idx_sphere_domain_links_user ON sphere_domain_links(user_id);

-- +goose Down
DROP TABLE IF EXISTS sphere_domain_links;
DROP TABLE IF EXISTS task_spheres;
