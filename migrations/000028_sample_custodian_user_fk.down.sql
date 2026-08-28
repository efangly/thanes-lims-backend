DROP INDEX IF EXISTS idx_samples_custodian_user_id;

ALTER TABLE samples DROP COLUMN custodian_user_id;

ALTER TABLE samples ADD COLUMN custodian VARCHAR(150) NOT NULL;
