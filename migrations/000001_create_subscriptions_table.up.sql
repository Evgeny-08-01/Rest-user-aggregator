CREATE TABLE IF NOT EXISTS subscriptions (
    id                SERIAL              PRIMARY KEY,
    service_name      VARCHAR(255)        NOT NULL,
    price             INTEGER             NOT NULL,
    user_id           UUID                NOT NULL,
    start_date        DATE                NOT NULL,
    end_date          DATE               
);

CREATE INDEX IF NOT EXISTS idx_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_service_name ON subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_dates ON subscriptions(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_dates ON subscriptions(user_id, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id, id);
