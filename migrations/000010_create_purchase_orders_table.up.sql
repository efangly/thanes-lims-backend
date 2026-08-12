CREATE TABLE purchase_orders (
    id         VARCHAR(20) PRIMARY KEY,
    item_id    VARCHAR(20)  NOT NULL REFERENCES inventory_items (id),
    quantity   INTEGER      NOT NULL,
    vendor     VARCHAR(200) NOT NULL,
    order_date TIMESTAMPTZ  NOT NULL,
    status     VARCHAR(20)  NOT NULL
);

CREATE INDEX idx_purchase_orders_item_id ON purchase_orders (item_id);
