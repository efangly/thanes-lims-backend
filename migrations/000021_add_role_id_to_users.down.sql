ALTER TABLE users
    ADD COLUMN role VARCHAR(20);

-- Reverse-populate role from role_id. 'Lab Manager' has no legacy
-- equivalent (it did not exist before this migration) so it falls back to
-- 'general' - the least-privileged legacy role - to keep the pre-existing
-- CHECK constraint satisfiable.
UPDATE users u
SET role = CASE r.name
    WHEN 'Admin' THEN 'admin'
    WHEN 'QA' THEN 'qa'
    WHEN 'Scientist' THEN 'scientist'
    ELSE 'general'
END
FROM roles r
WHERE r.id = u.role_id;

ALTER TABLE users
    ALTER COLUMN role SET NOT NULL;

ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'qa', 'scientist', 'general'));

DROP INDEX IF EXISTS idx_users_role_id;

ALTER TABLE users
    DROP COLUMN role_id;
