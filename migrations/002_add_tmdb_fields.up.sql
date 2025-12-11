-- Видаляємо старий constraint
ALTER TABLE movies DROP CONSTRAINT IF EXISTS movies_title_chat_id_unique;

CREATE TYPE content_type AS ENUM ('movie', 'series', 'anime', 'anime_series', 'cartoon', 'cartoon_series', 'documentary', 'documentary_series');

-- Додаємо нові колонки для TMDB даних
ALTER TABLE movies
ADD COLUMN tmdb_id INTEGER,
ADD COLUMN overview TEXT,
ADD COLUMN release_date VARCHAR(10),
ADD COLUMN rating DECIMAL(3,1),
ADD COLUMN poster_url VARCHAR(500),
ADD COLUMN genres JSONB,
ADD COLUMN content_type content_type DEFAULT 'movie';

-- Індекс для швидкого пошуку за tmdb_id
CREATE INDEX idx_movies_tmdb_id ON movies(tmdb_id);

-- Індекс для фільтрації по типу контенту
CREATE INDEX idx_movies_content_type ON movies(content_type);

-- Унікальність: якщо фільм з TMDB - не можна додати двічі (по tmdb_id)
CREATE UNIQUE INDEX idx_movies_chat_tmdb 
ON movies(chat_id, tmdb_id) 
WHERE tmdb_id IS NOT NULL;

-- Унікальність: якщо manual (без TMDB) - по назві
CREATE UNIQUE INDEX idx_movies_chat_title 
ON movies(chat_id, title) 
WHERE tmdb_id IS NULL;