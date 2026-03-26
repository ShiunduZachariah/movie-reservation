INSERT INTO users (clerk_id, email, name, role)
VALUES ('REPLACE_WITH_CLERK_ID', 'admin@yourdomain.com', 'System Admin', 'admin')
ON CONFLICT (clerk_id) DO NOTHING;

---- create above / drop below ----

DELETE FROM users WHERE clerk_id = 'REPLACE_WITH_CLERK_ID';

