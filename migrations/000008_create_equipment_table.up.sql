CREATE TABLE equipment (
    id                    VARCHAR(30) PRIMARY KEY,
    name                  VARCHAR(200) NOT NULL,
    type_code             VARCHAR(20)  NOT NULL,
    last_calibrated_at    TIMESTAMPTZ  NOT NULL,
    next_calibration_due  TIMESTAMPTZ  NOT NULL,
    usage_hours           INTEGER      NOT NULL DEFAULT 0
);
