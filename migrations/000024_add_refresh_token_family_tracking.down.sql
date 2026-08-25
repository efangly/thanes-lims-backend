DROP INDEX IF EXISTS idx_refresh_tokens_family_id;

ALTER TABLE refresh_tokens
    DROP COLUMN IF EXISTS family_id,
    DROP COLUMN IF EXISTS family_created_at,
    DROP COLUMN IF EXISTS user_agent,
    DROP COLUMN IF EXISTS ip_address;
