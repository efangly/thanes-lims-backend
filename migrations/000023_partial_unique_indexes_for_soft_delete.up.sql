-- A Retired record must not permanently block reuse of a value that was
-- unique to it (see ADR 0003 / CONTEXT.md "Retired"). Every plain UNIQUE
-- constraint/index on a now-soft-deletable table is replaced with a
-- partial unique index scoped to non-Retired rows.

-- users.email - the case ADR 0003 calls out by name.
ALTER TABLE users DROP CONSTRAINT users_email_key;
CREATE UNIQUE INDEX uq_users_email_active ON users (email) WHERE deleted_at IS NULL;

-- locations - sibling names and root (Cabinet) names must stay unique
-- only among non-Retired Locations.
DROP INDEX uq_locations_parent_name;
CREATE UNIQUE INDEX uq_locations_parent_name ON locations (parent_id, name) WHERE deleted_at IS NULL;

DROP INDEX uq_locations_root_name;
CREATE UNIQUE INDEX uq_locations_root_name ON locations (name) WHERE parent_id IS NULL AND deleted_at IS NULL;
