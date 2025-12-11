CREATE TABLE IF NOT EXISTS movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    chat_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT movies_title_chat_id_unique UNIQUE (title, chat_id)
);

CREATE INDEX IF NOT EXISTS idx_movies_chat_id ON movies(chat_id);