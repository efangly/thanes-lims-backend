DROP INDEX IF EXISTS uq_samples_barcode_id_active;
ALTER TABLE samples DROP COLUMN IF EXISTS description;
ALTER TABLE samples DROP COLUMN IF EXISTS barcode_id;
DELETE FROM id_sequences WHERE scope = 'sample_barcode';
