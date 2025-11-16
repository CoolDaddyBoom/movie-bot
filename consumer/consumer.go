package consumer

import (
	"context"
	"log"
	"time"

	"whattowatchbot/telegram"
)

type Consumer struct {
	client    *telegram.Client
	processor *telegram.Processor
	batchSize int
}

// New створює новий Consumer
func New(client *telegram.Client, processor *telegram.Processor, batchSize int) *Consumer {
	return &Consumer{
		client:    client,
		processor: processor,
		batchSize: batchSize,
	}
}

// Start запускає головний цикл бота
func (c *Consumer) Start(ctx context.Context) error {
	log.Println("Bot started, waiting for updates...")

	offset := 0

	for {
		// Перевірити чи не скасовано context (Ctrl+C)
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping...")
			return ctx.Err()
		default:
		}

		// Отримати updates
		updates, err := c.client.GetUpdates(offset, c.batchSize)
		if err != nil {
			log.Printf("Error getting updates: %v", err)
			time.Sleep(3 * time.Second) // Почекати перед retry
			continue
		}

		// Обробити кожен update
		for _, update := range updates {
			if err := c.processor.Process(ctx, update); err != nil {
				log.Printf("Error processing update %d: %v", update.UpdateID, err)
			}

			// Оновити offset
			offset = update.UpdateID + 1
		}

		// Якщо немає updates - почекати
		if len(updates) == 0 {
			time.Sleep(3 * time.Second)
		}
	}
}
