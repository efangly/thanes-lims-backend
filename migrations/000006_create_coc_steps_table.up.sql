CREATE TABLE coc_steps (
    id          BIGSERIAL PRIMARY KEY,
    sample_id   VARCHAR(30) NOT NULL REFERENCES samples (id),
    state       VARCHAR(10) NOT NULL,
    icon        VARCHAR(10) NOT NULL,
    title       VARCHAR(200) NOT NULL,
    meta        VARCHAR(255),
    who         VARCHAR(150) NOT NULL,
    occurred_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_coc_steps_sample_id ON coc_steps (sample_id, occurred_at);
