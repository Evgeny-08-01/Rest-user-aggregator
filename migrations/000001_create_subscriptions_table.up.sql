CREATE TABLE IF NOT EXISTS subscriptions (
    id                SERIAL              PRIMARY KEY,
    service_name      VARCHAR(255)        NOT NULL,
    price             INTEGER             NOT NULL,
    user_id           UUID                NOT NULL,
    start_date        TEXT                NOT NULL,
    end_date          TEXT               
);

CREATE INDEX IF NOT EXISTS idx_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_service_name ON subscriptions(service_name); 