-- migrations/000001_create_urls_table.down.sql
-- Откат создания таблицы урлов
DROP INDEX IF EXISTS idx_urls_short_url;
DROP TABLE IF EXISTS movies; 