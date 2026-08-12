CREATE TABLE inventory_items (
    id       VARCHAR(20) PRIMARY KEY,
    name     VARCHAR(200) NOT NULL,
    category VARCHAR(100) NOT NULL,
    quantity INTEGER      NOT NULL DEFAULT 0,
    unit     VARCHAR(30)  NOT NULL,
    min      INTEGER      NOT NULL DEFAULT 0,
    max      INTEGER      NOT NULL DEFAULT 0
);
