CREATE TABLE gauges (
    location  VARCHAR(100) PRIMARY KEY,
    unit      VARCHAR(20)      NOT NULL,
    range_min DOUBLE PRECISION NOT NULL,
    range_max DOUBLE PRECISION NOT NULL
);

CREATE TABLE sensor_readings (
    id          BIGSERIAL PRIMARY KEY,
    location    VARCHAR(100)     NOT NULL REFERENCES gauges (location),
    value       DOUBLE PRECISION NOT NULL,
    recorded_at TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX idx_sensor_readings_location_recorded_at ON sensor_readings (location, recorded_at DESC);

CREATE TABLE env_alerts (
    id           BIGSERIAL PRIMARY KEY,
    location     VARCHAR(100) NOT NULL,
    level        VARCHAR(10)  NOT NULL,
    title        VARCHAR(200) NOT NULL,
    message      VARCHAR(500) NOT NULL,
    triggered_at TIMESTAMPTZ  NOT NULL,
    resolved_at  TIMESTAMPTZ
);

CREATE INDEX idx_env_alerts_location_open ON env_alerts (location) WHERE resolved_at IS NULL;
