DROP INDEX IF EXISTS idx_equipment_location_id;
DROP INDEX IF EXISTS idx_equipment_vendor_id;
DROP INDEX IF EXISTS uq_equipment_serial_number_active;

ALTER TABLE equipment
    DROP COLUMN IF EXISTS location_id,
    DROP COLUMN IF EXISTS vendor_id,
    DROP COLUMN IF EXISTS installation_date,
    DROP COLUMN IF EXISTS model,
    DROP COLUMN IF EXISTS manufacturer,
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS serial_number;
