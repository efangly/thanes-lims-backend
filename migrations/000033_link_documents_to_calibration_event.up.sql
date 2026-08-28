-- Phase 6: a Document may optionally link to one CalibrationEvent (e.g. a
-- calibration certificate). Single-owner nullable FK, same pattern as the
-- Phase 5 equipment_id link.
ALTER TABLE documents
    ADD COLUMN calibration_event_id BIGINT REFERENCES calibration_events(id);

CREATE INDEX idx_documents_calibration_event_id
    ON documents (calibration_event_id);
