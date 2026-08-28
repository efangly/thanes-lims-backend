-- Phase 4: Sample Registry - optional scan Barcode ID + free-text Description.
-- See CONTEXT.md "Barcode ID". BarcodeID is separate from the Sample's ID
-- (which stays the SMP-{year}-{seq} sequence) and is optional.

ALTER TABLE samples ADD COLUMN barcode_id VARCHAR(30);
ALTER TABLE samples ADD COLUMN description TEXT NOT NULL DEFAULT '';

-- Unique across non-Retired Samples when set (partial index, matching the
-- soft-delete convention used for locations.barcode_code in 000027).
CREATE UNIQUE INDEX uq_samples_barcode_id_active
    ON samples (barcode_id)
    WHERE barcode_id IS NOT NULL AND deleted_at IS NULL;
