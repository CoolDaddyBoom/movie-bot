package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"whattowatchbot/internal/clients/tmdb"
	"whattowatchbot/internal/storage"
)

func (p *Processor) handleStart(ctx context.Context, chatID int) error {
	p.logger.Info("User started bot", "chat_id", chatID)

	text := `👋 Hello! I'm your BaoBaoMovie bot. 

	Send me a movie title to add it to your list! 
🎬
			
	Use /help to see all commands.`

	return p.client.SendMessage(chatID, text)
}

func (p *Processor) handleHelp(ctx context.Context, chatID int) error {
	text := `📖 Help:

/start - starts this bot
/help - shows this message
/random - gets a random movie from your list
/list - shows all movies in your list
/remove + title - removes a movie (don't write the + sign)

To add a movie, just send me its title! 🎬`

	return p.client.SendMessage(chatID, text)
}

func (p *Processor) handleRandom(ctx context.Context, chatID int) error {
	sharedChatID := normalizeUserChatID(chatID)

	p.logger.Debug("Getting random movie", "chat_id", chatID)

	movie, err := p.storage.PickRandom(ctx, sharedChatID)
	if err != nil {
		p.logger.Error("Failed to pick random movie", "chat_id", chatID, "error", err)
		return p.client.SendMessage(chatID, "❌ Failed to get random movie. Try again later.")
	}

	if movie == nil {
		p.logger.Debug("No movies found", "chat_id", chatID)
		text := `You don't have any saved movies yet!
		Add some by sending me their titles! 
		🎬`
		return p.client.SendMessage(chatID, text)
	}

	p.logger.Info("Picked random movie", "chat_id", chatID, "title", movie.Title)

	// Просто показуємо фільм БЕЗ видалення
	text := p.formatRandomMovie(movie)

	return p.client.SendPhotoOrMessage(chatID, movie.PosterURL, text)
}

func (p *Processor) handleList(ctx context.Context, chatID int) error {
	sharedChatID := normalizeUserChatID(chatID)

	p.logger.Debug("Processing list movies", "chat_id", chatID)

	movies, err := p.storage.List(ctx, sharedChatID)
	if err != nil {
		p.logger.Error("Failed to list movies", "error", err, "chat_id", chatID)
		return p.client.SendMessage(chatID, "❌ Failed to get your list. Try again later.")
	}

	if len(movies) == 0 {
		p.logger.Debug("No movies in list", "chat_id", sharedChatID)
		text := `You don't have any saved movies yet! 
		Add some by sending me their titles! 
		🎬`
		return p.client.SendMessage(chatID, text)
	}

	p.logger.Info("Listed movies", "chat_id", chatID, "count", len(movies))

	text := p.formatMovieList(movies)

	return p.client.SendMessage(chatID, text)
}

func (p *Processor) handleAddMovie(ctx context.Context, chatID int, title string) error {
	sharedChatID := normalizeUserChatID(chatID)

	if strings.HasPrefix(title, "/") && title == "/remove" {
		return p.client.SendMessage(chatID, "❌ Please provide a movie title after the \"/remove\" command.")
	} else if strings.HasPrefix(title, "/") {
		return p.client.SendMessage(chatID, "❌ Unknown command. Use /help for the list of commands.")
	}

	// Нормалізувати назву
	normalizedTitle := normalizeTitle(title)

	p.logger.Debug("Adding movie", "chat_id", chatID, "original_title", title)

	// Перевірити чи не порожня після очищення
	if normalizedTitle == "" {
		return p.client.SendMessage(chatID, "❌ Movie title cannot be empty")
	}

	// Check if the title is too long
	if len(normalizedTitle) > 200 {
		return p.client.SendMessage(chatID, "❌ Movie title is too long (maximum 200 characters)")
	}

	// Search movie in TMDB
	results, err := p.tmdbClient.SearchMulti(ctx, normalizedTitle)
	if errors.Is(err, tmdb.ErrMovieNotFound) {
		p.logger.Info("Movie not found in TMDB", "chat_id", sharedChatID, "title", normalizedTitle)
		text := fmt.Sprintf(`⚠️ Movie "%s" not found in TMDB database.
							Would you like to add it manually (without TMDB info)?
							Reply with: /add_manual %s`, normalizedTitle, normalizedTitle)

		return p.client.SendMessage(chatID, text)
	}

	// Інша помилка (API недоступний) (точно???)
	if err != nil {
		p.logger.Error("TMDB API error", "error", err, "title", normalizedTitle)
		return p.client.SendMessage(chatID, "❌ Failed to search movie. Try again later.")
	}

	// Беремо перший результат (поки що)
	result := results[0]

	p.logger.Info("Movie found in TMDB",
		"chat_id", sharedChatID,
		"title", result.Title,
		"tmdb_id", result.ID)

	movie := p.multiResultToStorageMovie(&result, sharedChatID)

	// Перевірити чи не існує
	exists, err := p.storage.IsExists(ctx, movie)
	if err != nil {
		p.logger.Error("Failed to check if movie exists", "error", err, "chat_id", sharedChatID)
		return p.client.SendMessage(chatID, "❌ Failed to add movie. Try again later.")
	}

	if exists {
		p.logger.Debug("Movie already exists", "chat_id", chatID, "title", normalizedTitle)
		text := fmt.Sprintf("ℹ️ The movie \"%s\" is already in your list!", normalizedTitle)
		return p.client.SendMessage(chatID, text)
	}

	// Save
	if err := p.storage.Save(ctx, movie); err != nil {
		p.logger.Error("Failed to save movie", "error", err, "chat_id", sharedChatID, "title", normalizedTitle)
		return p.client.SendMessage(chatID, "❌ Failed to add movie. Try again later.")
	}

	p.logger.Info("Movie added", "chat_id", sharedChatID, "title", normalizedTitle)

	// Форматуємо відповідь
	text := p.formatMovieAdded(movie)

	return p.client.SendPhotoOrMessage(chatID, movie.PosterURL, text)
}

func (p *Processor) handleRemove(ctx context.Context, chatID int, title string) error {
	sharedChatID := normalizeUserChatID(chatID)

	normalizedTitle := normalizeTitle(title)

	if normalizedTitle == "" {
		return p.client.SendMessage(chatID, "❌ Please specify the movie title after the /remove command")
	}

	p.logger.Debug("Removing movie", "chat_id", chatID, "title", normalizedTitle)

	movie := &storage.Movie{
		Title:  normalizedTitle,
		ChatID: sharedChatID,
	}

	// Check if it exists
	exists, err := p.storage.IsExists(ctx, movie)
	if err != nil {
		p.logger.Error("Failed to check if movie exists", "error", err, "chat_id", sharedChatID)
		return p.client.SendMessage(chatID, "❌ Failed to remove movie. Try again later.")
	}

	if !exists {
		p.logger.Debug("Movie not found for removal", "chat_id", chatID, "title", normalizedTitle)
		return p.client.SendMessage(chatID,
			fmt.Sprintf("❌ The movie \"%s\" was not found in your list", normalizedTitle))
	}

	// Remove
	if err := p.storage.Remove(ctx, movie); err != nil {
		p.logger.Error("Failed to remove movie", "error", err, "chat_id", sharedChatID, "title", normalizedTitle)
		return p.client.SendMessage(chatID, "❌ Failed to remove movie. Try again later.")
	}

	p.logger.Info("Movie removed", "chat_id", sharedChatID, "title", normalizedTitle)
	return p.client.SendMessage(chatID,
		fmt.Sprintf("✅ The movie \"%s\" has been removed from your list", normalizedTitle))
}
