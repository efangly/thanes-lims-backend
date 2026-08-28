DROP INDEX IF EXISTS uq_locations_barcode_active;
ALTER TABLE locations DROP COLUMN IF EXISTS barcode_code;
DELETE FROM id_sequences WHERE scope = 'location_barcode';

-- Drop any equipment_storage rows before narrowing the constraints back -
-- they cannot exist under the original schema.
DELETE FROM locations WHERE kind = 'equipment_storage';

ALTER TABLE locations DROP CONSTRAINT locations_level_type_check;
ALTER TABLE locations
    ADD CONSTRAINT locations_level_type_check
        CHECK (level_type IN ('cabinet', 'shelf', 'slot', 'sub_slot'));

ALTER TABLE locations DROP COLUMN IF EXISTS kind;
