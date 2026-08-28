-- Phase 7: asset fields on InventoryItem (task.md / CONTEXT.md#inventory).
-- custodian_user_id FKs users and is NOT NULL (the User responsible for the
-- item - CONTEXT.md "Custodian"); no production InventoryItem data exists
-- yet, so the column is added NOT NULL outright rather than backfilled - any
-- pre-existing rows would fail the add, the intended signal that a real
-- backfill strategy is needed first.
-- manufacturer is a plain descriptive string, distinct from vendor_id which
-- FKs the Vendor master record (CONTEXT.md#vendors). location_id references
-- a Location of Kind equipment_storage (ADR 0007) - not enforced in SQL,
-- checked in the use case. default_vendor (free text, drives auto-reorder)
-- is deliberately kept alongside vendor_id.
ALTER TABLE inventory_items
    ADD COLUMN custodian_user_id BIGINT NOT NULL REFERENCES users (id),
    ADD COLUMN manufacturer      VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN vendor_id         VARCHAR(30) REFERENCES vendors (id),
    ADD COLUMN location_id       VARCHAR(30) REFERENCES locations (id);

CREATE INDEX idx_inventory_items_custodian_user_id ON inventory_items (custodian_user_id);
CREATE INDEX idx_inventory_items_vendor_id ON inventory_items (vendor_id);
CREATE INDEX idx_inventory_items_location_id ON inventory_items (location_id);
