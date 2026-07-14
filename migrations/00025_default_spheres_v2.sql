-- +goose Up
INSERT INTO life_spheres (id, user_id, name, sort_order, created_at, updated_at)
SELECT gen_random_uuid(), u.id, d.name, d.sort_order, now(), now()
FROM users u
CROSS JOIN (
    VALUES
        ('Деньги', 0),
        ('Карьера GO', 1),
        ('Цех ЧПУ', 2),
        ('Дом и быт', 3),
        ('Хобби и отдых', 4)
) AS d(name, sort_order)
WHERE NOT EXISTS (
    SELECT 1 FROM life_spheres ls WHERE ls.user_id = u.id AND ls.name = d.name
);

-- +goose Down
DELETE FROM life_spheres
WHERE name IN ('Деньги', 'Карьера GO', 'Цех ЧПУ', 'Дом и быт', 'Хобби и отдых');
