package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"whattowatchbot/storage"

	_ "github.com/mattn/go-sqlite3" // Драйвер SQLite (blank import)
)

type Storage struct {
	db *sql.DB // З'єднання з SQLite БД
}

var ErrNoRows = errors.New("sql: no rows in result set")

func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Init(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS movies (
				title TEXT NOT NULL,
				chat_id INTEGER NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (title, chat_id)
			  )`

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (s *Storage) Save(ctx context.Context, m *storage.Movie) error {

	if m.Title == "" {
		return errors.New("title cannot be empty")
	}

	query := `INSERT INTO movies (title, chat_id) VALUES (?, ?)`

	_, err := s.db.ExecContext(ctx, query, m.Title, m.ChatID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("movie '%s' already exists", m.Title)
		}
		return fmt.Errorf("failed to insert movie: %w", err)
	}

	return nil
}

func (s *Storage) PickRandom(ctx context.Context, chatID int) (*storage.Movie, error) {
	query := `SELECT title, chat_id FROM movies 
              WHERE chat_id = ? 
              ORDER BY RANDOM() 
              LIMIT 1`

	var movie storage.Movie
	err := s.db.QueryRowContext(ctx, query, chatID).Scan(&movie.Title, &movie.ChatID)

	if err == sql.ErrNoRows {
		return nil, nil // Немає фільмів - не помилка
	}
	if err != nil {
		return nil, fmt.Errorf("can't pick random movie: %w", err)
	}

	return &movie, nil
}

func (s *Storage) Remove(ctx context.Context, m *storage.Movie) error {
	query := `DELETE FROM movies 
	          WHERE title = ? AND chat_id = ?`

	_, err := s.db.ExecContext(ctx, query, m.Title, m.ChatID)
	if err != nil {
		return fmt.Errorf("failed to delete the movie: %w", err)
	}
	return nil
}

func (s *Storage) List(ctx context.Context, chatID int) ([]*storage.Movie, error) {
	q := `SELECT title, chat_id FROM movies WHERE chat_id = ? ORDER BY title ASC`

	// Query замість QueryRow - бо багато рядків!
	rows, err := s.db.QueryContext(ctx, q, chatID)
	if err != nil {
		return nil, fmt.Errorf("can't get movies: %w", err)
	}
	defer rows.Close() // ОБОВ'ЯЗКОВО закрити!

	// Створити порожній slice
	var movies []*storage.Movie

	// Цикл по всіх рядках
	for rows.Next() {
		var m storage.Movie

		// Прочитати один рядок
		if err := rows.Scan(&m.Title, &m.ChatID); err != nil {
			return nil, fmt.Errorf("can't scan movie: %w", err)
		}

		// Додати до slice
		movies = append(movies, &m)
	}

	// Перевірити чи не було помилки під час ітерації
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return movies, nil
}

func (s *Storage) IsExists(ctx context.Context, m *storage.Movie) (bool, error) {
	query := `SELECT COUNT(*) FROM movies 
	          WHERE title = ? AND chat_id = ?`

	var count int

	err := s.db.QueryRowContext(ctx, query, m.Title, m.ChatID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check movie existence: %w", err)
	}

	return count > 0, nil
}
