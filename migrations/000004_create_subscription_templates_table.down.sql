-- ============================================================
-- ФАЙЛ: migrations/000004_create_subscription_templates_table.down.sql
-- НАЗНАЧЕНИЕ: откат таблицы шаблонов подписок
-- ============================================================

DROP TABLE IF EXISTS subscription_templates CASCADE;
