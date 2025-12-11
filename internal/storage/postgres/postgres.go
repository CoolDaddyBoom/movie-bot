package postgres

import (
	"context"
	"errors"
	"fmt"
	"whattowatchbot/internal/storage"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Storage struct {
	db *gorm.DB
}

func New(dsn string) (*Storage, error) {

	config := gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Логувати всі SQL запити
	}

	db, err := gorm.Open(postgres.Open(dsn), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Init(ctx context.Context) error {
	// AutoMigrate створює/оновлює таблиці на основі структури
	err := s.db.WithContext(ctx).AutoMigrate(&storage.Movie{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}

func (s *Storage) Save(ctx context.Context, m *storage.Movie) error {
	if m.Title == "" {
		return errors.New("title cannot be empty")
	}

	result := s.db.WithContext(ctx).Create(m)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("movie '%s' already exists", m.Title)
		}
		return fmt.Errorf("failed to save movie: %w", result.Error)
	}

	return nil
}

func (s *Storage) PickRandom(ctx context.Context, chatID int) (*storage.Movie, error) {
	var movie storage.Movie

	// ORDER BY RANDOM() - PostgreSQL специфічна функція
	result := s.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("RANDOM()").
		First(&movie)

	// Перевірка чи знайдено
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Немає фільмів - не помилка
		}
		return nil, fmt.Errorf("failed to pick random movie: %w", result.Error)
	}

	return &movie, nil
}

func (s *Storage) Remove(ctx context.Context, m *storage.Movie) error {

	result := s.db.WithContext(ctx).
		Where("title = ? AND chat_id = ?", m.Title, m.ChatID).
		Delete(&storage.Movie{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete the movie: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("movie not found")
	}

	return nil
}

func (s *Storage) List(ctx context.Context, chatID int) ([]*storage.Movie, error) {
	var movies []*storage.Movie

	result := s.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("title ASC").
		Find(&movies)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list movies: %w", result.Error)
	}

	return movies, nil
}

func (s *Storage) IsExists(ctx context.Context, m *storage.Movie) (bool, error) {
	var count int64

	result := s.db.WithContext(ctx).
		Model(&storage.Movie{}).
		Where("title = ? AND chat_id = ?", m.Title, m.ChatID).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check movie existence: %w", result.Error)
	}

	return count > 0, nil
}

func (s *Storage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
