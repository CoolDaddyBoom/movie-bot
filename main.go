package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"whattowatchbot/consumer"
	"whattowatchbot/storage/sqlite"
	"whattowatchbot/telegram"
)

func main() {
	// Завантажити .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Отримати токен
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is not set")
	}

	// Отримати шлях до БД
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "movies.db" // За замовчуванням
	}

	// Створити storage
	storage, err := sqlite.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	// Ініціалізувати БД (створити таблиці)
	ctx := context.Background()
	if err := storage.Init(ctx); err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	log.Println("Database initialized successfully")

	// Створити Telegram client
	client := telegram.NewClient(token)

	// Створити processor
	processor := telegram.NewProcessor(client, storage)

	// Створити consumer
	cons := consumer.New(client, processor, 5) // 5 updates за раз

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Слухати Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	// Запустити бота
	log.Println("Starting bot...")
	if err := cons.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Bot error: %v", err)
	}

	log.Println("Bot stopped gracefully")
}
