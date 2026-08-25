-- Retired (soft delete) support for every audited Module per ADR 0003 -
-- a delete must leave the record (and its Audit Trail) queryable, so a
-- hard DELETE is replaced with GORM's gorm.DeletedAt convention (a NULL
-- deleted_at means "active"; a delete stamps the current time instead of
-- removing the row).
--
-- coc_steps and calibration_events are deliberately excluded: both are
-- append-only sub-resources (their repositories expose Append/AppendStep
-- but no Delete), so there is no delete path to make safe yet.

ALTER TABLE samples ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_samples_deleted_at ON samples (deleted_at);

ALTER TABLE locations ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_locations_deleted_at ON locations (deleted_at);

ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

ALTER TABLE test_results ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_test_results_deleted_at ON test_results (deleted_at);

ALTER TABLE equipment ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_equipment_deleted_at ON equipment (deleted_at);

ALTER TABLE inventory_items ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_inventory_items_deleted_at ON inventory_items (deleted_at);

ALTER TABLE purchase_orders ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_purchase_orders_deleted_at ON purchase_orders (deleted_at);

ALTER TABLE documents ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_documents_deleted_at ON documents (deleted_at);
