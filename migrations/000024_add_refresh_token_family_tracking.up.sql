ALTER TABLE refresh_tokens
    ADD COLUMN family_id TEXT,
    ADD COLUMN family_created_at TIMESTAMPTZ,
    ADD COLUMN user_agent TEXT,
    ADD COLUMN ip_address TEXT;

-- Backfill existing rows as their own single-token family. These predate
-- rotation chains entirely, so treating each as its own family (rather than
-- trying to reconstruct real chains) is safe - they are all either already
-- revoked or expired in practice given the 7-day refresh token TTL.
UPDATE refresh_tokens
SET family_id = id::text,
    family_created_at = created_at
WHERE family_id IS NULL;

ALTER TABLE refresh_tokens
    ALTER COLUMN family_id SET NOT NULL,
    ALTER COLUMN family_created_at SET NOT NULL;

CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens (family_id);
