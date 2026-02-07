CREATE TABLE IF NOT EXISTS urls_users (
    url_id INTEGER REFERENCES urls(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (url_id, user_id)
);