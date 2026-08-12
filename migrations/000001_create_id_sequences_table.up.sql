CREATE TABLE id_sequences (
    scope   VARCHAR(50) NOT NULL,
    year    INTEGER     NOT NULL DEFAULT 0,
    current BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (scope, year)
);
