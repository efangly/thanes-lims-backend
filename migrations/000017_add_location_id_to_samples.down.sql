DROP INDEX uq_samples_active_location;

ALTER TABLE samples
    ADD COLUMN location VARCHAR(150) NOT NULL DEFAULT '';

ALTER TABLE samples
    DROP COLUMN location_id;
