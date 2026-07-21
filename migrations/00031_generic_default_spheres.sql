-- +goose Up
-- Generic product defaults: rename personal seed names if still present.
UPDATE life_spheres
SET name = 'Карьера', updated_at = now()
WHERE name = 'Карьера GO';

UPDATE life_spheres
SET name = 'Здоровье', updated_at = now()
WHERE name = 'Цех ЧПУ'
  AND NOT EXISTS (
      SELECT 1 FROM life_spheres s2
      WHERE s2.user_id = life_spheres.user_id AND s2.name = 'Здоровье'
  );

-- Users who already have "Здоровье" keep "Цех ЧПУ" as-is (avoid unique clash);
-- rename leftover personal shop name to generic work label.
UPDATE life_spheres
SET name = 'Работа', updated_at = now()
WHERE name = 'Цех ЧПУ';

INSERT INTO life_spheres (id, user_id, name, sort_order, created_at, updated_at)
SELECT gen_random_uuid(), u.id, d.name, d.sort_order, now(), now()
FROM users u
CROSS JOIN (
    VALUES
        ('Деньги', 0),
        ('Карьера', 1),
        ('Здоровье', 2),
        ('Дом и быт', 3),
        ('Хобби и отдых', 4)
) AS d(name, sort_order)
WHERE NOT EXISTS (
    SELECT 1 FROM life_spheres ls WHERE ls.user_id = u.id AND ls.name = d.name
);

-- +goose Down
UPDATE life_spheres SET name = 'Карьера GO', updated_at = now() WHERE name = 'Карьера';
UPDATE life_spheres SET name = 'Цех ЧПУ', updated_at = now() WHERE name IN ('Здоровье', 'Работа');
