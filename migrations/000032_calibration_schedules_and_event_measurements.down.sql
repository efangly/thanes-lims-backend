ALTER TABLE calibration_events
    DROP COLUMN IF EXISTS result,
    DROP COLUMN IF EXISTS acceptance_value,
    DROP COLUMN IF EXISTS calibrate_value,
    DROP COLUMN IF EXISTS calibration_type;

DROP TABLE IF EXISTS calibration_schedules;
