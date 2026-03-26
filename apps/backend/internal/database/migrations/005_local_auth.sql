ALTER TABLE users
ADD COLUMN IF NOT EXISTS password_hash TEXT;

INSERT INTO users (clerk_id, email, name, role, password_hash)
VALUES (
  'local:shiunduzachariah@gmail.com',
  'shiunduzachariah@gmail.com',
  'TestUser',
  'admin',
  crypt('Test123', gen_salt('bf'))
)
ON CONFLICT (email) DO UPDATE
SET
  clerk_id = EXCLUDED.clerk_id,
  name = EXCLUDED.name,
  role = EXCLUDED.role,
  password_hash = EXCLUDED.password_hash,
  updated_at = NOW();

---- create above / drop below ----

DELETE FROM users WHERE email = 'shiunduzachariah@gmail.com';

ALTER TABLE users
DROP COLUMN IF EXISTS password_hash;
