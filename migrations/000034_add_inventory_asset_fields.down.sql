DROP INDEX IF EXISTS idx_inventory_items_location_id;
DROP INDEX IF EXISTS idx_inventory_items_vendor_id;
DROP INDEX IF EXISTS idx_inventory_items_custodian_user_id;

ALTER TABLE inventory_items
    DROP COLUMN IF EXISTS location_id,
    DROP COLUMN IF EXISTS vendor_id,
    DROP COLUMN IF EXISTS manufacturer,
    DROP COLUMN IF EXISTS custodian_user_id;
