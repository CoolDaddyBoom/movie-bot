package telegram

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"whattowatchbot/storage"
)

// normalizeTitle приводить назву до стандартного вигляду для збереження
func normalizeTitle(title string) string {
	// 1. Видалити пробіли з країв
	title = strings.TrimSpace(title)

	// 2. Видалити лапки з початку і кінця
	title = strings.Trim(title, `"`)

	// 3. Знову видалити пробіли (якщо були після лапок)
	title = strings.TrimSpace(title)

	// 4. Привести до Title Case (Перша Велика, Решта Малі)
	return toTitleCase(title)
}

// toTitleCase перетворює "слово слово" в "Слово Слово"
func toTitleCase(s string) string {
	if s == "" {
		return s
	}

	words := strings.Fields(s) // Розбити на слова
	for i, word := range words {
		if len(word) > 0 {
			// Перша літера велика, решта малі
			runes := []rune(strings.ToLower(word))
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}

	return strings.Join(words, " ")
}

type Processor struct {
	client  *Client
	storage storage.Storage
}

func NewProcessor(client *Client, storage storage.Storage) *Processor {
	return &Processor{
		client:  client,
		storage: storage,
	}
}

func (p *Processor) Process(ctx context.Context, upd Update) error {
	if upd.Message == nil {
		return nil
	}

	chatID := upd.Message.Chat.ID
	text := upd.Message.Text
	username := upd.Message.From.Username

	switch {
	case text == "/start":
		return p.handleStart(ctx, chatID)
	case text == "/help":
		return p.handleHelp(ctx, chatID)
	case text == "/random":
		return p.handleRandom(ctx, chatID)
	case text == "/list":
		return p.handleList(ctx, chatID)
	case strings.HasPrefix(text, "/remove "): // ← Додати
		title := strings.TrimPrefix(text, "/remove ")
		return p.handleRemove(ctx, chatID, title)
	default:
		// Додати фільм
		return p.handleAddMovie(ctx, chatID, username, text)
	}
}

func (p *Processor) handleStart(ctx context.Context, chatID int) error {
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

	movie, err := p.storage.PickRandom(ctx, sharedChatID)
	if err != nil {
		return err
	}

	if movie == nil {
		text := `You don't have any saved movies yet!
		Add some by sending me their titles! 
		🎬`
		return p.client.SendMessage(chatID, text)
	}

	// Просто показуємо фільм БЕЗ видалення
	text := fmt.Sprintf("🎬 %s\n\nTo remove it from the list after watching, send:\n/remove %s",
		movie.Title,
		movie.Title)

	return p.client.SendMessage(chatID, text)
	// ← НЕ видаляємо!
}

func (p *Processor) handleList(ctx context.Context, chatID int) error {
	sharedChatID := normalizeUserChatID(chatID)

	movies, err := p.storage.List(ctx, sharedChatID)
	if err != nil {
		return err
	}

	if len(movies) == 0 {
		text := `You don't have any saved movies yet! 
		Add some by sending me their titles! 
		🎬`
		return p.client.SendMessage(chatID, text)
	}

	text := fmt.Sprintf("📋 You have %d movies:\n\n", len(movies))
	for i, movie := range movies {
		text += fmt.Sprintf("%d. %s\n", i+1, movie.Title)
	}

	return p.client.SendMessage(chatID, text)
}

func (p *Processor) handleAddMovie(ctx context.Context, chatID int, username, title string) error {
	sharedChatID := normalizeUserChatID(chatID)

	if strings.HasPrefix(title, "/") && title == "/remove" {
		return p.client.SendMessage(chatID, "❌ Please provide a movie title after the \"/remove\" command.")
	} else if strings.HasPrefix(title, "/") {
		return p.client.SendMessage(chatID, "❌ Unknown command. Use /help for the list of commands.")
	}

	// Нормалізувати назву
	normalizedTitle := normalizeTitle(title)

	// Перевірити чи не порожня після очищення
	if normalizedTitle == "" {
		return p.client.SendMessage(chatID, "❌ Movie title cannot be empty")
	}

	// Check if the title is too long
	if len(normalizedTitle) > 200 {
		return p.client.SendMessage(chatID, "❌ Movie title is too long (maximum 200 characters)")
	}

	// Створити movie
	movie := &storage.Movie{
		Title:  normalizedTitle,
		ChatID: sharedChatID,
	}

	// Перевірити чи не існує
	exists, err := p.storage.IsExists(ctx, movie)
	if err != nil {
		return err
	}

	if exists {
		text := fmt.Sprintf("ℹ️ The movie \"%s\" is already in your list!", normalizedTitle)
		return p.client.SendMessage(chatID, text)
	}

	// Save
	if err := p.storage.Save(ctx, movie); err != nil {
		return err
	}

	text := fmt.Sprintf("✅ The movie \"%s\" has been added to your list!", normalizedTitle)
	return p.client.SendMessage(chatID, text)
}

func (p *Processor) handleRemove(ctx context.Context, chatID int, title string) error {
	sharedChatID := normalizeUserChatID(chatID)

	normalizedTitle := normalizeTitle(title)

	if normalizedTitle == "" {
		return p.client.SendMessage(chatID, "❌ Please specify the movie title after the /remove command")
	}

	movie := &storage.Movie{
		Title:  normalizedTitle,
		ChatID: sharedChatID,
	}

	// Check if it exists
	exists, err := p.storage.IsExists(ctx, movie)
	if err != nil {
		return err
	}

	if !exists {
		return p.client.SendMessage(chatID,
			fmt.Sprintf("❌ The movie \"%s\" was not found in your list", normalizedTitle))
	}

	// Remove
	if err := p.storage.Remove(ctx, movie); err != nil {
		return err
	}

	return p.client.SendMessage(chatID,
		fmt.Sprintf("✅ The movie \"%s\" has been removed from your list", normalizedTitle))
}

func normalizeUserChatID(chatID int) int {
	// Список особливих користувачів (ти і дівчина)
	specialUsers := map[int]int{
		613544049:  613544049, // Твій chat_id → твій же (зміни на свій!)
		7465672598: 613544049, // Її chat_id → твій chat_id (зміни на її!)
	}

	// Якщо користувач особливий - повертає спільний ID
	if sharedID, exists := specialUsers[chatID]; exists {
		return sharedID
	}

	// Звичайний користувач - повертає свій ID
	return chatID
}
