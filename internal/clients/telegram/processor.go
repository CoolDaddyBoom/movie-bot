package telegram

import (
	"context"
	"log/slog"
	"strings"
	"whattowatchbot/internal/clients/tmdb"
	"whattowatchbot/internal/storage"
)

type Processor struct {
	client     *Client
	tmdbClient *tmdb.Client
	storage    storage.Storage
	logger     *slog.Logger
}

func NewProcessor(client *Client, tmdbClient *tmdb.Client, storage storage.Storage, logger *slog.Logger) *Processor {
	return &Processor{
		client:     client,
		tmdbClient: tmdbClient,
		storage:    storage,
		logger:     logger,
	}
}

func (p *Processor) Process(ctx context.Context, upd Update) error {
	if upd.Message == nil {
		return nil
	}

	chatID := upd.Message.Chat.ID
	text := upd.Message.Text

	p.logger.Debug("Received message",
		"chat_id", chatID,
		"text", text,
		"user", upd.Message.From.Username)

	switch {
	case text == "/start":
		return p.handleStart(ctx, chatID)
	case text == "/help":
		return p.handleHelp(ctx, chatID)
	case text == "/random":
		return p.handleRandom(ctx, chatID)
	case text == "/list":
		return p.handleList(ctx, chatID)
	case strings.HasPrefix(text, "/remove "):
		title := strings.TrimPrefix(text, "/remove ")
		return p.handleRemove(ctx, chatID, title)
	default:
		// Додати фільм
		return p.handleAddMovie(ctx, chatID, text)
	}
}
