CREATE TABLE locations (
    id         VARCHAR(30) PRIMARY KEY,
    parent_id  VARCHAR(30) REFERENCES locations (id),
    name       VARCHAR(100) NOT NULL,
    level_type VARCHAR(20)  NOT NULL CHECK (level_type IN ('cabinet', 'shelf', 'slot', 'sub_slot')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_locations_parent_id ON locations (parent_id);

-- Siblings must have distinct names.
CREATE UNIQUE INDEX uq_locations_parent_name ON locations (parent_id, name);

-- parent_id IS NULL marks a Cabinet (root); Postgres does not treat NULLs as
-- equal in a regular unique index, so root name uniqueness needs its own
-- partial index.
CREATE UNIQUE INDEX uq_locations_root_name ON locations (name) WHERE parent_id IS NULL;
