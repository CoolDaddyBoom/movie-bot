package logger

import (
	"log/slog"
	"os"

	"whattowatchbot/internal/config"
)

// New створює налаштований logger на основі конфігурації
func NewLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: parseLevel(cfg.App.LogLevel),
	}

	if cfg.App.Environment == "production" {
		// JSON для production (легко парсити)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text для development (читабельніше)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// parseLevel конвертує string в slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
