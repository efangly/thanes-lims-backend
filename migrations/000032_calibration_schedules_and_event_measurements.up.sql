-- Phase 6: recurring calibration schedules + measurement fields on the
-- append-only calibration event log (task.md Phase 6, CONTEXT.md
-- "Calibration Schedule").

CREATE TABLE calibration_schedules (
    id              BIGSERIAL PRIMARY KEY,
    equipment_id    VARCHAR(30) NOT NULL REFERENCES equipment(id),
    label           VARCHAR(200) NOT NULL,
    next_due_date   TIMESTAMPTZ NOT NULL,
    interval_months INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_calibration_schedules_equipment_id
    ON calibration_schedules (equipment_id)
    WHERE deleted_at IS NULL;

-- Measurement fields recorded per calibration. All nullable / default ''
-- so existing rows keep working. calibration_type, when it matches a
-- schedule's label, drives that schedule's auto-advance.
ALTER TABLE calibration_events
    ADD COLUMN calibration_type VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN calibrate_value  VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN acceptance_value VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN result           VARCHAR(10)  NOT NULL DEFAULT '';
