INSERT INTO genres (name) VALUES
  ('Action'), ('Comedy'), ('Drama'), ('Horror'),
  ('Romance'), ('Sci-Fi'), ('Thriller'), ('Animation')
ON CONFLICT (name) DO NOTHING;

---- create above / drop below ----

DELETE FROM genres WHERE name IN
  ('Action', 'Comedy', 'Drama', 'Horror', 'Romance', 'Sci-Fi', 'Thriller', 'Animation');

