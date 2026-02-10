CREATE TABLE IF NOT EXISTS urls_users (
    url_id INTEGER,
    user_id INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (url_id, user_id)
);
-- REFERENCES urls(id) ON DELETE CASCADE
-- INTEGER REFERENCES users(id) ON DELETE CASCADE