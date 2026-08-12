CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    actor_id    BIGINT,
    actor_role  VARCHAR(20),
    method      VARCHAR(10)  NOT NULL,
    path        VARCHAR(255) NOT NULL,
    resource    VARCHAR(100),
    resource_id VARCHAR(100),
    status_code INTEGER      NOT NULL,
    metadata    JSONB,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs (actor_id);
