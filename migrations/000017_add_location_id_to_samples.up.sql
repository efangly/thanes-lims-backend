ALTER TABLE samples
    ADD COLUMN location_id VARCHAR(30) REFERENCES locations (id);

ALTER TABLE samples
    DROP COLUMN location;

-- A leaf Location may hold at most one Sample that's still occupying it.
-- "Occupying" means status pending/testing/completed; transferred frees the
-- Location back up. See CONTEXT.md#storage-location.
CREATE UNIQUE INDEX uq_samples_active_location ON samples (location_id)
    WHERE location_id IS NOT NULL AND status IN ('pending', 'testing', 'completed');
