package tmdb

import "errors"

var (
	ErrMovieNotFound = errors.New("movie not found in TMDB")
	ErrAPIError      = errors.New("TMDB API error")
	ErrInvalidAPIKey = errors.New("invalid TMDB API key")
)
