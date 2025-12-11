package tmdb

// MultiSearchResult - результат з /search/multi
type MultiSearchResult struct {
	MediaType string `json:"media_type"` // "movie", "tv", або "person"

	// Для movie
	ID          int    `json:"id"`
	Title       string `json:"title"`        // Для movie
	ReleaseDate string `json:"release_date"` // Для movie

	// Для TV
	Name         string `json:"name"`           // Для TV
	FirstAirDate string `json:"first_air_date"` // Для TV

	// Спільні поля
	Overview      string   `json:"overview"`
	PosterPath    string   `json:"poster_path"`
	VoteAverage   float64  `json:"vote_average"`
	GenreIDs      []int    `json:"genre_ids"`      // В multi search - тільки ID
	OriginCountry []string `json:"origin_country"` // Для визначення аніме
}

// MultiSearchResponse - відповідь від /search/multi
type MultiSearchResponse struct {
	Results      []MultiSearchResult `json:"results"`
	TotalResults int                 `json:"total_results"`
	Page         int                 `json:"page"`
}

// GetTitle повертає назву (movie.title або tv.name)
func (r *MultiSearchResult) GetTitle() string {
	if r.MediaType == "movie" {
		return r.Title
	}
	return r.Name
}

// GetReleaseDate повертає дату (movie.release_date або tv.first_air_date)
func (r *MultiSearchResult) GetReleaseDate() string {
	if r.MediaType == "movie" {
		return r.ReleaseDate
	}
	return r.FirstAirDate
}

// IsAnime визначає чи це аніме
func (r *MultiSearchResult) IsAnime() bool {
	for _, country := range r.OriginCountry {
		if country == "JP" {
			return true
		}
	}
	return false
}

// IsPerson перевіряє чи це персона (актор/режисер)
func (r *MultiSearchResult) IsPerson() bool {
	return r.MediaType == "person"
}
