-- +goose Up
CREATE TABLE life_spheres (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(128) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT life_sphere_name_check CHECK (length(trim(name)) > 0),
    UNIQUE (user_id, name)
);

CREATE INDEX life_spheres_user_sort_idx ON life_spheres (user_id, sort_order, created_at);

INSERT INTO life_spheres (id, user_id, name, sort_order, created_at, updated_at)
SELECT gen_random_uuid(), u.id, d.name, d.sort_order, now(), now()
FROM users u
CROSS JOIN (
    VALUES
        ('Карьера', 0),
        ('Личная жизнь', 1),
        ('Здоровье', 2),
        ('Финансы', 3),
        ('Обучение', 4)
) AS d(name, sort_order)
WHERE NOT EXISTS (
    SELECT 1 FROM life_spheres ls WHERE ls.user_id = u.id
);

-- +goose Down
DROP TABLE IF EXISTS life_spheres;
