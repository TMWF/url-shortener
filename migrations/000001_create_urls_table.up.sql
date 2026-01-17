-- migrations/000001_create_urls_table.up.sql
-- Создание таблицы урлов
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(8) NOT NULL,
    original_url VARCHAR(250) NOT NULL
);

-- Базовый индекс для поиска по названию
CREATE INDEX IF NOT EXISTS idx_urls_short_url ON urls(short_url);