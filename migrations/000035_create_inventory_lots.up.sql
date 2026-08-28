-- Phase 8: InventoryLot + receiving stock (task.md / CONTEXT.md "Inventory
-- Lot"). An InventoryItem's on-hand quantity stops being a directly-written
-- column and becomes the derived sum of its lots' quantities.
CREATE TABLE inventory_lots (
    id          VARCHAR(30) PRIMARY KEY,
    item_id     VARCHAR(30) NOT NULL REFERENCES inventory_items (id),
    lot_no      VARCHAR(100) NOT NULL,
    expire_date DATE,
    quantity    INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inventory_lots_item_id ON inventory_lots (item_id);
-- One lot number per item: receiving more of an existing lot adds to it
-- rather than creating a duplicate row.
CREATE UNIQUE INDEX idx_inventory_lots_item_lot_no ON inventory_lots (item_id, lot_no);

-- Preserve each item's current on-hand quantity as an opening lot so the
-- derived sum matches what the column held, then drop the now-derived
-- column. No production data exists yet, but this keeps the migration safe
-- and the down migration lossless.
INSERT INTO inventory_lots (id, item_id, lot_no, expire_date, quantity)
SELECT 'LOT-OPEN-' || substr(md5(id), 1, 8), id, 'OPENING', NULL, quantity
FROM inventory_items
WHERE quantity <> 0;

ALTER TABLE inventory_items DROP COLUMN quantity;
