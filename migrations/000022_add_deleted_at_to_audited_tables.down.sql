DROP INDEX IF EXISTS idx_documents_deleted_at;
ALTER TABLE documents DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_purchase_orders_deleted_at;
ALTER TABLE purchase_orders DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_inventory_items_deleted_at;
ALTER TABLE inventory_items DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_equipment_deleted_at;
ALTER TABLE equipment DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_test_results_deleted_at;
ALTER TABLE test_results DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_users_deleted_at;
ALTER TABLE users DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_locations_deleted_at;
ALTER TABLE locations DROP COLUMN deleted_at;

DROP INDEX IF EXISTS idx_samples_deleted_at;
ALTER TABLE samples DROP COLUMN deleted_at;
