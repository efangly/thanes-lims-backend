CREATE TABLE documents (
    id           VARCHAR(20) PRIMARY KEY,
    name         VARCHAR(200) NOT NULL,
    type         VARCHAR(20)  NOT NULL,
    version      VARCHAR(10)  NOT NULL,
    created_by   VARCHAR(150) NOT NULL,
    issued_at    TIMESTAMPTZ  NOT NULL,
    access_level VARCHAR(50)  NOT NULL DEFAULT '',
    locked       BOOLEAN      NOT NULL DEFAULT false,
    storage_key  VARCHAR(500) NOT NULL
);

CREATE TABLE doc_history (
    id          BIGSERIAL PRIMARY KEY,
    document_id VARCHAR(20)  NOT NULL REFERENCES documents (id),
    version     VARCHAR(10)  NOT NULL,
    change      VARCHAR(500) NOT NULL,
    date        TIMESTAMPTZ  NOT NULL,
    who         VARCHAR(150) NOT NULL
);

CREATE INDEX idx_doc_history_document_id ON doc_history (document_id, date);
