-- Відкочуємо зміни у зворотному порядку

-- Видаляємо індекси
DROP INDEX IF EXISTS idx_movies_chat_title;
DROP INDEX IF EXISTS idx_movies_chat_tmdb;
DROP INDEX IF EXISTS idx_movies_content_type;
DROP INDEX IF EXISTS idx_movies_tmdb_id;

-- Видаляємо колонки
ALTER TABLE movies
DROP COLUMN IF EXISTS content_type,
DROP COLUMN IF EXISTS genres,
DROP COLUMN IF EXISTS poster_url,
DROP COLUMN IF EXISTS rating,
DROP COLUMN IF EXISTS release_date,
DROP COLUMN IF EXISTS overview,
DROP COLUMN IF EXISTS tmdb_id;

-- Видаляємо enum тип
DROP TYPE IF EXISTS content_type;