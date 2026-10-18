-- +goose Up

INSERT INTO units (id, name) VALUES
    (1, 'шт'),
    (2, 'кг'),
    (3, 'г'),
    (4, 'л'),
    (5, 'мл')
ON CONFLICT (id) DO NOTHING;

-- +goose Down

DELETE FROM units WHERE id IN (1, 2, 3, 4, 5);