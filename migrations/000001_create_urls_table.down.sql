-- migrations/000001_create_urls_table.down.sql
-- Откат создания таблицы урлов
DROP INDEX IF EXISTS idx_urls_original_url_unique;
DROP INDEX IF EXISTS idx_urls_short_url_unique;
DROP TABLE IF EXISTS urls; 