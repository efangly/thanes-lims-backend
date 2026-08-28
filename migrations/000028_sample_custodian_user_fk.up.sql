-- Phase 3: Sample.Custodian (free text) -> CustodianUserID (FK to users).
-- See CONTEXT.md "Custodian". No production Sample data exists yet, so the
-- old free-text column is dropped outright rather than backfilled - any
-- pre-existing rows would fail the NOT NULL add, which is the intended
-- signal that a real backfill strategy is needed first.

ALTER TABLE samples DROP COLUMN custodian;

ALTER TABLE samples
    ADD COLUMN custodian_user_id BIGINT NOT NULL REFERENCES users (id);

CREATE INDEX idx_samples_custodian_user_id ON samples (custodian_user_id);
