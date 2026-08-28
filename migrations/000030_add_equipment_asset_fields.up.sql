-- Phase 5: descriptive/asset fields on Equipment (task.md). All nullable -
-- existing rows keep working untouched. Category is kept separate from the
-- existing type_code (type_code stays short and drives the ID sequence;
-- category is the human-facing grouping). manufacturer is a plain string,
-- distinct from vendor_id which FKs the Vendor master record
-- (CONTEXT.md#vendors). location_id references a Location of Kind
-- equipment_storage (ADR 0007) - not enforced in SQL, checked in the use
-- case.
ALTER TABLE equipment
    ADD COLUMN serial_number     VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN category          VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN manufacturer      VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN model             VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN installation_date TIMESTAMPTZ,
    ADD COLUMN vendor_id         VARCHAR(30) REFERENCES vendors(id),
    ADD COLUMN location_id       VARCHAR(30) REFERENCES locations(id);

-- serial_number is unique across active (non-Retired) Equipment only, and
-- only when set (ADR 0003 / CONTEXT.md "Retired").
CREATE UNIQUE INDEX uq_equipment_serial_number_active
    ON equipment (serial_number)
    WHERE serial_number <> '' AND deleted_at IS NULL;

CREATE INDEX idx_equipment_vendor_id ON equipment (vendor_id);
CREATE INDEX idx_equipment_location_id ON equipment (location_id);
