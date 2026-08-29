CREATE TABLE IF NOT EXISTS subscriptions (
    id                SERIAL                 PRIMARY KEY,
    user_id           UUID                   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id       INTEGER REFERENCES     subscription_templates(id) ON DELETE SET NULL,
    start_date        DATE                   NOT NULL,
    end_date          DATE,   
    UNIQUE (user_id, template_id)
);             
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_template_id ON subscriptions(template_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_dates ON subscriptions(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_dates ON subscriptions(user_id, start_date, end_date);
