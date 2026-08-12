CREATE TABLE samples (
    id          VARCHAR(30) PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    type        VARCHAR(20)  NOT NULL,
    custodian   VARCHAR(150) NOT NULL,
    location    VARCHAR(150) NOT NULL,
    status      VARCHAR(20)  NOT NULL,
    received_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_samples_status ON samples (status);
