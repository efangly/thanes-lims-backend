DROP INDEX IF EXISTS idx_documents_equipment_id;
ALTER TABLE documents DROP COLUMN IF EXISTS equipment_id;
