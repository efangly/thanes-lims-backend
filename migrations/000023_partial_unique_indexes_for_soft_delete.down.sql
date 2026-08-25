DROP INDEX IF EXISTS uq_locations_root_name;
CREATE UNIQUE INDEX uq_locations_root_name ON locations (name) WHERE parent_id IS NULL;

DROP INDEX IF EXISTS uq_locations_parent_name;
CREATE UNIQUE INDEX uq_locations_parent_name ON locations (parent_id, name);

DROP INDEX IF EXISTS uq_users_email_active;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
