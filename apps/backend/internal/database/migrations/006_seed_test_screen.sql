WITH seeded_screen AS (
    INSERT INTO screens (name, total_seats)
    VALUES ('Test Screen', 10)
    ON CONFLICT (name) DO UPDATE
    SET total_seats = EXCLUDED.total_seats,
        updated_at = NOW()
    RETURNING id
)
INSERT INTO seats (screen_id, row, number)
SELECT s.id, 'A', n
FROM seeded_screen s
CROSS JOIN generate_series(1, 10) AS n
ON CONFLICT (screen_id, row, number) DO NOTHING;

---- create above / drop below ----

DELETE FROM seats
WHERE screen_id IN (SELECT id FROM screens WHERE name = 'Test Screen');

DELETE FROM screens WHERE name = 'Test Screen';
