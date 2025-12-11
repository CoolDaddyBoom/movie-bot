package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Bot      BotConfig
	TMDB     TMDBConfig
	App      AppConfig
}

type DatabaseConfig struct {
	DSN            string `env:"DATABASE_DSN" env-required:"true"`
	MaxConnections int    `env:"DATABASE_MAX_CONNECTIONS" env-default:"10"`
	ConnTimeout    int    `env:"DATABASE_CONN_TIMEOUT" env-default:"30"` // in seconds
}

type TMDBConfig struct {
	APIKey   string `env:"TMDB_API_KEY" env-required:"true"`
	Language string `env:"TMDB_LANGUAGE" env-default:"en-US"`
}

type BotConfig struct {
	Token string `env:"BOT_TOKEN" env-required:"true"`
}

type AppConfig struct {
	Environment string `env:"APP_ENVIRONMENT" env-default:"development"`
	LogLevel    string `env:"APP_LOG_LEVEL" env-default:"info"`
}

func Load() (*Config, error) {
	// Завантажити .env файл
	_ = godotenv.Load() // Ігноруємо помилку якщо файлу немає
	var cfg Config

	// cleanenv автоматично:
	// 1. Завантажить .env
	// 2. Прочитає env variables
	// 3. Валідує required поля
	// 4. Встановить defaults
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}
