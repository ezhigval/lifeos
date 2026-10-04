-- +goose Up
-- 00035 ошибочно ссылался на несуществующую таблицу spheres(id); сферы живут в life_spheres.
ALTER TABLE task_spheres
    DROP CONSTRAINT IF EXISTS task_spheres_sphere_id_fkey,
    ADD CONSTRAINT task_spheres_sphere_id_fkey
        FOREIGN KEY (sphere_id) REFERENCES life_spheres(id) ON DELETE CASCADE;

ALTER TABLE sphere_domain_links
    DROP CONSTRAINT IF EXISTS sphere_domain_links_sphere_id_fkey,
    ADD CONSTRAINT sphere_domain_links_sphere_id_fkey
        FOREIGN KEY (sphere_id) REFERENCES life_spheres(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE task_spheres
    DROP CONSTRAINT IF EXISTS task_spheres_sphere_id_fkey;
ALTER TABLE sphere_domain_links
    DROP CONSTRAINT IF EXISTS sphere_domain_links_sphere_id_fkey;
