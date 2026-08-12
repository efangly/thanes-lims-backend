CREATE TABLE test_results (
    id        VARCHAR(20) PRIMARY KEY,
    sample_id VARCHAR(30)  NOT NULL REFERENCES samples (id),
    test_name VARCHAR(200) NOT NULL,
    analyst   VARCHAR(150) NOT NULL,
    result    VARCHAR(200) NOT NULL DEFAULT '',
    flag      VARCHAR(5)   NOT NULL DEFAULT 'ok',
    ref_range VARCHAR(200) NOT NULL DEFAULT '',
    status    VARCHAR(25)  NOT NULL
);

CREATE INDEX idx_test_results_sample_id ON test_results (sample_id);
CREATE INDEX idx_test_results_status ON test_results (status);
