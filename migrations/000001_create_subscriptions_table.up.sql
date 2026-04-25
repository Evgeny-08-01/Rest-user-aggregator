CREATE TABLE IF NOT EXISTS subscriptions (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    service_name      TEXT                NOT NULL,
    price             INTEGER             NOT NULL,
    user_id           INTEGER             NOT NULL,
    start_date        TEXT                NOT NULL,
    end_date          TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_service_name ON subscriptions(service_name);