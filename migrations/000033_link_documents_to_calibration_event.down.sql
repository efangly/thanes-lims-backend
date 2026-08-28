DROP INDEX IF EXISTS idx_documents_calibration_event_id;

ALTER TABLE documents
    DROP COLUMN IF EXISTS calibration_event_id;
