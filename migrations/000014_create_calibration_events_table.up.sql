CREATE TABLE calibration_events (
    id                    BIGSERIAL PRIMARY KEY,
    equipment_id          VARCHAR(30) NOT NULL REFERENCES equipment (id),
    calibrated_at         TIMESTAMPTZ NOT NULL,
    next_calibration_due  TIMESTAMPTZ NOT NULL,
    performed_by          VARCHAR(150) NOT NULL,
    notes                 VARCHAR(255),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_calibration_events_equipment_id ON calibration_events (equipment_id, calibrated_at);
