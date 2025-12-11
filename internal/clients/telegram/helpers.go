package telegram

import (
	"fmt"
	"strings"
	"unicode"
	"whattowatchbot/internal/clients/tmdb"
	"whattowatchbot/internal/storage"
)

// normalizeTitle приводить назву до стандартного вигляду для збереження
func normalizeTitle(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}

	words := strings.Fields(s)
	for i, w := range words {
		runes := []rune(w)
		for j := range runes {
			if j == 0 {
				runes[j] = unicode.ToUpper(runes[j])
			} else {
				runes[j] = unicode.ToLower(runes[j])
			}
		}
		words[i] = string(runes)
	}

	return strings.Join(words, " ")
}

func normalizeUserChatID(chatID int) int {
	// Список особливих користувачів
	specialUsers := map[int]int{
		613544049:  613544049,
		7465672598: 613544049,
	}

	// Якщо користувач особливий - повертає спільний ID
	if sharedID, exists := specialUsers[chatID]; exists {
		return sharedID
	}

	// Звичайний користувач - повертає свій ID
	return chatID
}

// multiResultToStorageMovie конвертує TMDB multi search result в модель БД
func (p *Processor) multiResultToStorageMovie(result *tmdb.MultiSearchResult, chatID int) *storage.Movie {
	// Poster URL
	posterURL := ""
	if result.PosterPath != "" {
		posterURL = p.tmdbClient.GetPosterURL(result.PosterPath, "w500")
	}

	// Визначити ContentType
	contentType := p.determineContentType(result)

	// Genres - в multi search тільки IDs, конвертуємо
	genres := p.genreIDsToNames(result.GenreIDs)

	return &storage.Movie{
		Title:       normalizeTitle(result.GetTitle()),
		ChatID:      chatID,
		TMDBID:      result.ID,
		Overview:    result.Overview,
		ReleaseDate: result.GetReleaseDate(),
		Rating:      result.VoteAverage,
		PosterURL:   posterURL,
		Genres:      genres,
		ContentType: contentType,
	}
}

// determineContentType визначає тип контенту
func (p *Processor) determineContentType(result *tmdb.MultiSearchResult) storage.ContentType {
	isAnime := result.IsAnime()

	if result.MediaType == "movie" {
		if isAnime {
			return storage.ContentTypeAnime
		}
		return storage.ContentTypeMovie
	}

	if result.MediaType == "tv" {
		if isAnime {
			return storage.ContentTypeAnimeSeries
		}
		return storage.ContentTypeSeries
	}

	// Fallback
	return storage.ContentTypeMovie
}

// genreIDsToNames конвертує genre IDs в назви
func (p *Processor) genreIDsToNames(genreIDs []int) []string {
	// Простий mapping найпопулярніших жанрів
	genreMap := map[int]string{
		28:    "Action",
		12:    "Adventure",
		16:    "Animation",
		35:    "Comedy",
		80:    "Crime",
		99:    "Documentary",
		18:    "Drama",
		10751: "Family",
		14:    "Fantasy",
		36:    "History",
		27:    "Horror",
		10402: "Music",
		9648:  "Mystery",
		10749: "Romance",
		878:   "Science Fiction",
		10770: "TV Movie",
		53:    "Thriller",
		10752: "War",
		37:    "Western",

		// TV жанри
		10759: "Action & Adventure",
		10762: "Kids",
		10763: "News",
		10764: "Reality",
		10765: "Sci-Fi & Fantasy",
		10766: "Soap",
		10767: "Talk",
		10768: "War & Politics",
	}

	names := make([]string, 0, len(genreIDs))
	for _, id := range genreIDs {
		if name, ok := genreMap[id]; ok {
			names = append(names, name)
		}
	}

	return names
}

// formatMovieAdded форматує повідомлення про доданий фільм/серіал
func (p *Processor) formatMovieAdded(movie *storage.Movie) string {
	// Емодзі по типу
	emoji := "🎬"
	typeName := "Movie"

	switch movie.ContentType {
	case storage.ContentTypeSeries:
		emoji = "📺"
		typeName = "Series"
	case storage.ContentTypeAnime:
		emoji = "🗾"
		typeName = "Anime"
	case storage.ContentTypeAnimeSeries:
		emoji = "🍙"
		typeName = "Anime Series"
	}

	// Рік
	year := ""
	if len(movie.ReleaseDate) >= 4 {
		year = movie.ReleaseDate[:4]
	}

	// Жанри
	genres := ""
	if len(movie.Genres) > 0 {
		genres = "\n🎭 " + movie.GenresString()
	}

	// Рейтинг
	rating := ""
	if movie.Rating >= 0 {
		rating = fmt.Sprintf("\n⭐ %.1f/10", movie.Rating)
	}

	return fmt.Sprintf("%s <b>%s</b> (%s)\n📌 %s%s%s",
		emoji,
		movie.Title,
		year,
		typeName,
		rating,
		genres)
}

func (p *Processor) formatRandomMovie(movie *storage.Movie) string {
	// Емодзі по типу
	emoji := "🎬"
	typeName := "Movie"

	switch movie.ContentType {
	case storage.ContentTypeSeries:
		emoji = "📺"
		typeName = "Series"
	case storage.ContentTypeAnime:
		emoji = "🗾"
		typeName = "Anime"
	case storage.ContentTypeAnimeSeries:
		emoji = "🍙"
		typeName = "Anime Series"
	}

	// Рік
	year := ""
	if len(movie.ReleaseDate) >= 4 {
		year = movie.ReleaseDate[:4]
	}

	// Жанри
	genres := ""
	if len(movie.Genres) > 0 {
		genres = "\n🎭 " + movie.GenresString()
	}

	// Рейтинг
	rating := ""
	if movie.Rating >= 0 {
		rating = fmt.Sprintf("\n⭐ %.1f/10", movie.Rating)
	}

	// Опис (обрізати)
	overview := ""
	if movie.Overview != "" {
		overview = "\n\n" + truncate(movie.Overview, 300)
	}

	return fmt.Sprintf("🎲 <b>Random pick!</b>\n\n%s <b>%s</b> (%s)\n📌 %s%s%s%s",
		emoji,
		movie.Title,
		year,
		typeName,
		rating,
		genres,
		overview)
}

// formatMovieList форматує список фільмів
func (p *Processor) formatMovieList(movies []*storage.Movie) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📋 <b>Your list (%d items):</b>\n\n", len(movies)))

	for i, movie := range movies {

		// Рік
		year := ""
		if len(movie.ReleaseDate) >= 4 {
			year = fmt.Sprintf(" (%s)", movie.ReleaseDate[:4])
		}

		rating := ""
		if movie.Rating >= 0 {
			rating = fmt.Sprintf(" ⭐%.1f/10", movie.Rating)
		}

		sb.WriteString(fmt.Sprintf("%d. <b>%s</b> %s %s\n", i+1, movie.Title, year, rating))
	}

	return sb.String()
}

// truncate обрізає текст до maxLen символів
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
