-- ADR-0009: a Box is a Location (level_type = 'box') that holds many samples
-- by Cell position, instead of the one-sample-per-leaf rule. A Box carries a
-- Grid (rows x cols) and hangs off a Shelf/Slot/Sub-slot; it never has child
-- Locations. A Cell is not a node - it is samples.position ('A1', 'H12').

-- 1. 'box' joins the level_type vocabulary. It is a terminal marker, not a
--    fixed depth (see docs/adr/0009).
ALTER TABLE locations DROP CONSTRAINT locations_level_type_check;
ALTER TABLE locations
    ADD CONSTRAINT locations_level_type_check
        CHECK (level_type IN (
            'cabinet', 'shelf', 'slot', 'sub_slot',
            'building', 'room', 'zone', 'box'
        ));

-- 2. Grid columns - non-null exactly for boxes. rows is capped at 26 (Cell
--    rows are named A..Z), cols at 99 (two-digit Cell columns).
ALTER TABLE locations ADD COLUMN rows SMALLINT;
ALTER TABLE locations ADD COLUMN cols SMALLINT;
ALTER TABLE locations
    ADD CONSTRAINT locations_grid_check
        CHECK (
            (level_type = 'box'
                AND rows BETWEEN 1 AND 26
                AND cols BETWEEN 1 AND 99)
            OR
            (level_type <> 'box' AND rows IS NULL AND cols IS NULL)
        );

-- 3. A sample's Cell. Null for every sample not in a box, including all
--    existing rows - no data migration (ADR-0009).
ALTER TABLE samples ADD COLUMN position VARCHAR(4);

-- 4. The pre-ADR-0009 leaf-occupancy index keyed on location_id alone, which
--    a box breaks: many active samples share one box's location_id. Narrow it
--    to non-box put-away (position IS NULL)...
DROP INDEX uq_samples_active_location;
CREATE UNIQUE INDEX uq_samples_active_location
    ON samples (location_id)
    WHERE location_id IS NOT NULL
      AND position IS NULL
      AND status IN ('pending', 'testing', 'completed');

-- 5. ...and add the box-cell equivalent: one active sample per
--    (location_id, position). The application layer checks this too, but the
--    partial unique index makes a batch Move-within-box atomic - a position
--    clash fails the whole transaction.
CREATE UNIQUE INDEX uq_samples_box_cell_active
    ON samples (location_id, position)
    WHERE position IS NOT NULL
      AND deleted_at IS NULL
      AND status IN ('pending', 'testing', 'completed');
