-- User Status (see CONTEXT.md "## User Lifecycle" and ADR 0010): a reversible
-- active/suspended flag, orthogonal to Retired (deleted_at) and to Role. Every
-- existing User is active.
ALTER TABLE users
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended'));

CREATE INDEX idx_users_status ON users (status);
