-- ============================================================
-- ФАЙЛ: migrations/000004_create_subscription_templates_table.up.sql
-- НАЗНАЧЕНИЕ: создание таблицы шаблонов подписок
-- ============================================================

CREATE TABLE IF NOT EXISTS subscription_templates (
    id SERIAL PRIMARY KEY,
    service_name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0)
);

CREATE UNIQUE INDEX idx_unique_service_name_ci 
ON subscription_templates (LOWER(TRIM(service_name)));

COMMENT ON TABLE subscription_templates IS 'Шаблоны подписок, создаваемые админом';
COMMENT ON COLUMN subscription_templates.service_name IS 'Название сервиса (уникальное, без учёта регистра и пробелов)';
COMMENT ON COLUMN subscription_templates.price IS 'Цена подписки в рублях (целое число)';
