-- Phase 2: generalize Location into two trees (see docs/adr/0007 and
-- CONTEXT.md "Location Kind"), and give every node a scan Barcode.

-- 1. Kind discriminator. Every existing row is the original Sample tree.
ALTER TABLE locations
    ADD COLUMN kind VARCHAR(20) NOT NULL DEFAULT 'sample_storage'
        CHECK (kind IN ('sample_storage', 'equipment_storage'));

-- Keep the default so callers that don't set kind still land in the Sample
-- tree (matches the application-layer default).

-- 2. equipment_storage introduces three new level types (building > room >
--    zone > cabinet > shelf); "cabinet" and "shelf" are reused from the
--    Sample hierarchy at a different depth, disambiguated by kind.
ALTER TABLE locations DROP CONSTRAINT locations_level_type_check;
ALTER TABLE locations
    ADD CONSTRAINT locations_level_type_check
        CHECK (level_type IN (
            'cabinet', 'shelf', 'slot', 'sub_slot',
            'building', 'room', 'zone'
        ));

-- 3. Location Barcode: auto-generated per node, unique across non-Retired
--    Locations. Backfill every existing node with a sequential code and
--    advance the id_sequences counter so the app continues from there.
ALTER TABLE locations ADD COLUMN barcode_code VARCHAR(20);

WITH numbered AS (
    SELECT id, row_number() OVER (ORDER BY id) AS rn FROM locations
)
UPDATE locations l
SET barcode_code = 'LOC-BC-' || lpad(numbered.rn::text, 5, '0')
FROM numbered
WHERE numbered.id = l.id;

INSERT INTO id_sequences (scope, year, current)
VALUES ('location_barcode', 0, (SELECT count(*) FROM locations))
ON CONFLICT (scope, year) DO UPDATE SET current = EXCLUDED.current;

CREATE UNIQUE INDEX uq_locations_barcode_active
    ON locations (barcode_code)
    WHERE barcode_code IS NOT NULL AND deleted_at IS NULL;
