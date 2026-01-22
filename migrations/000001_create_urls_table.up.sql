-- migrations/000001_create_urls_table.up.sql
-- Создание таблицы урлов
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(8) NOT NULL,
    original_url VARCHAR(250) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_short_url_unique ON urls(short_url);
CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original_url_unique ON urls (original_url);