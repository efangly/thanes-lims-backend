CREATE TABLE roles (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The 5 system Roles - see CONTEXT.md#access-control and ADR 0002.
INSERT INTO roles (name) VALUES
    ('Admin'),
    ('Lab Manager'),
    ('QA'),
    ('Scientist'),
    ('General');
