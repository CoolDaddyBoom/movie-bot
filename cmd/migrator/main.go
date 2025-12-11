package main

import (
	"fmt"
	"log"
	"os"

	"whattowatchbot/internal/config"
	"whattowatchbot/internal/logger"
	"whattowatchbot/internal/migrator"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.NewLogger(cfg)

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrator <command>")
		fmt.Println("Commands: up, down, status")
		os.Exit(1)
	}

	command := os.Args[1]

	migrator, err := migrator.New(cfg.Database.DSN, logger)
	if err != nil {
		logger.Error("Failed to connect", "error", err)
		os.Exit(1)
	}
	defer migrator.Close()

	switch command {
	case "up":
		if err := migrator.Up(); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}
		fmt.Println("✅ All migrations applied")
	case "down":
		if err := migrator.Down(); err != nil {
			log.Fatalf("❌ Rollback failed: %v", err)
		}
		fmt.Println("✅ Migration rolled back")
	case "status":
		if err := migrator.Status(); err != nil {
			log.Fatalf("❌ Status check failed: %v", err)
		}
	default:
		log.Fatalf("❌ Unknown command: %s", command)
	}
}
