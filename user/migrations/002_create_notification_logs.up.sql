CREATE TABLE IF NOT EXISTS notification_logs (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL,
    type       VARCHAR(50),
    subject    VARCHAR(255),
    body       TEXT,
    sent_at    TIMESTAMP NOT NULL DEFAULT now(),
    status     VARCHAR(20),
    error_msg  TEXT
    );