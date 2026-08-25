ALTER TABLE users
    ADD COLUMN role_id BIGINT REFERENCES roles (id);

-- Backfill role_id from the legacy role string before dropping it.
UPDATE users SET role_id = (SELECT id FROM roles WHERE roles.name = 'Admin') WHERE role = 'admin';
UPDATE users SET role_id = (SELECT id FROM roles WHERE roles.name = 'QA') WHERE role = 'qa';
UPDATE users SET role_id = (SELECT id FROM roles WHERE roles.name = 'Scientist') WHERE role = 'scientist';
UPDATE users SET role_id = (SELECT id FROM roles WHERE roles.name = 'General') WHERE role = 'general';

ALTER TABLE users
    ALTER COLUMN role_id SET NOT NULL;

CREATE INDEX idx_users_role_id ON users (role_id);

ALTER TABLE users
    DROP COLUMN role;
