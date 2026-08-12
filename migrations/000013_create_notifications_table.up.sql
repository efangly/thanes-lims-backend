CREATE TABLE notifications (
    id                VARCHAR(15) PRIMARY KEY,
    recipient_user_id BIGINT,
    tone              VARCHAR(10)  NOT NULL,
    icon              VARCHAR(30)  NOT NULL,
    title             VARCHAR(200) NOT NULL,
    message           VARCHAR(500) NOT NULL,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    read              BOOLEAN      NOT NULL DEFAULT false
);

CREATE INDEX idx_notifications_recipient ON notifications (recipient_user_id, read);
