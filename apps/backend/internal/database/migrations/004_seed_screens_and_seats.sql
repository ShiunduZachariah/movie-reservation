WITH inserted_screens AS (
    INSERT INTO screens (name, total_seats)
    VALUES ('Screen 1', 40), ('Screen 2', 40)
    ON CONFLICT (name) DO UPDATE SET total_seats = EXCLUDED.total_seats
    RETURNING id, name
)
INSERT INTO seats (screen_id, row, number)
SELECT s.id, rows.row_label, nums.n
FROM inserted_screens s
CROSS JOIN (VALUES ('A'), ('B'), ('C'), ('D'), ('E')) AS rows(row_label)
CROSS JOIN generate_series(1, 8) AS nums(n)
ON CONFLICT (screen_id, row, number) DO NOTHING;

---- create above / drop below ----

DELETE FROM seats
WHERE screen_id IN (SELECT id FROM screens WHERE name IN ('Screen 1', 'Screen 2'));

DELETE FROM screens WHERE name IN ('Screen 1', 'Screen 2');
