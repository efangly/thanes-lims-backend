CREATE TABLE vendors (
    id            VARCHAR(30) PRIMARY KEY,
    name          VARCHAR(200) NOT NULL,
    contact_name  VARCHAR(200) NOT NULL DEFAULT '',
    contact_phone VARCHAR(50)  NOT NULL DEFAULT '',
    contact_email VARCHAR(200) NOT NULL DEFAULT '',
    address       TEXT         NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

-- Vendor Name is unique across non-Retired rows only (ADR 0003 /
-- CONTEXT.md "Retired").
CREATE UNIQUE INDEX uq_vendors_name_active ON vendors (name) WHERE deleted_at IS NULL;
CREATE INDEX idx_vendors_deleted_at ON vendors (deleted_at);
