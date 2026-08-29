DROP INDEX IF EXISTS uq_samples_box_cell_active;

-- Restore the original leaf-occupancy index (keyed on location_id alone).
DROP INDEX IF EXISTS uq_samples_active_location;
CREATE UNIQUE INDEX uq_samples_active_location ON samples (location_id)
    WHERE location_id IS NOT NULL AND status IN ('pending', 'testing', 'completed');

ALTER TABLE samples DROP COLUMN IF EXISTS position;

-- Boxes cannot exist under the pre-ADR-0009 schema.
DELETE FROM locations WHERE level_type = 'box';

ALTER TABLE locations DROP CONSTRAINT IF EXISTS locations_grid_check;
ALTER TABLE locations DROP COLUMN IF EXISTS rows;
ALTER TABLE locations DROP COLUMN IF EXISTS cols;

ALTER TABLE locations DROP CONSTRAINT locations_level_type_check;
ALTER TABLE locations
    ADD CONSTRAINT locations_level_type_check
        CHECK (level_type IN (
            'cabinet', 'shelf', 'slot', 'sub_slot',
            'building', 'room', 'zone'
        ));
