package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"whattowatchbot/consumer"
	"whattowatchbot/internal/clients/telegram"
	"whattowatchbot/internal/clients/tmdb"
	"whattowatchbot/internal/config"
	"whattowatchbot/internal/logger"
	"whattowatchbot/internal/storage/postgres"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log := logger.NewLogger(cfg)

	// Створюємо PostgreSQL storage
	storage, err := postgres.New(cfg.Database.DSN)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer storage.Close()

	// Ініціалізуємо БД (міграції)
	ctx := context.Background()
	if err := storage.Init(ctx); err != nil {
		log.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	log.Info("✅ Database connected and initialized")

	// Створюємо компоненти бота
	// TMDB клієнт
	tmdbClient := tmdb.NewClient(cfg.TMDB.APIKey, cfg.TMDB.Language, log)
	// Telegram
	client := telegram.NewClient(cfg.Bot.Token, log)
	processor := telegram.NewProcessor(client, tmdbClient, storage, log)

	// Consumer
	cons := consumer.New(client, processor, 5)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обробка Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("\n⚠️  Shutting down gracefully...")
		cancel()
	}()

	// Запускаємо бота
	log.Info("Bot started", "environment", cfg.App.Environment)
	if err := cons.Start(ctx); err != nil && err != context.Canceled {
		log.Error("Bot error", "error", err)
		os.Exit(1)
	}

	log.Info("👋 Bot stopped gracefully")
}
