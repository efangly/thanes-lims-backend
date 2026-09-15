CREATE TABLE partner_devices (
    serial   VARCHAR(100) PRIMARY KEY,
    location VARCHAR(100) NOT NULL REFERENCES gauges (location),
    active   BOOLEAN      NOT NULL DEFAULT true
);

CREATE INDEX idx_partner_devices_location ON partner_devices (location);
