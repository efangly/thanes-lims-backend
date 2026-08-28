ALTER TABLE inventory_items ADD COLUMN quantity INTEGER NOT NULL DEFAULT 0;

UPDATE inventory_items i
SET quantity = COALESCE(
    (SELECT SUM(l.quantity) FROM inventory_lots l WHERE l.item_id = i.id), 0
);

DROP TABLE inventory_lots;
