-- Phase 5: a Document may optionally link to one Equipment (e.g. a
-- warranty). Single-owner nullable FK rather than a join table - the
-- Document module has no other cross-links today and a warranty/manual
-- belongs to exactly one machine (task.md Phase 5 decision).
ALTER TABLE documents
    ADD COLUMN equipment_id VARCHAR(30) REFERENCES equipment(id);

CREATE INDEX idx_documents_equipment_id ON documents (equipment_id);
