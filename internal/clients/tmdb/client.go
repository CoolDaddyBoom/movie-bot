package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"time"
)

const (
	tmdbScheme = "https"
	tmdbHost   = "api.themoviedb.org"
	tmdbAPIv3  = "/3"

	posterScheme = "https"
	posterHost   = "image.tmdb.org"
	posterBase   = "/t/p"
)

type Client struct {
	apiKey     string
	language   string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(apiKey string, language string, logger *slog.Logger) *Client {
	return &Client{
		apiKey:     apiKey,
		language:   language,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

// SearchMulti шукає фільми, серіали та людей одночасно
func (c *Client) SearchMulti(ctx context.Context, query string) ([]MultiSearchResult, error) {
	c.logger.Debug("Searching in TMDB (multi)", "query", query)

	params := url.Values{}
	params.Add("api_key", c.apiKey)
	params.Add("language", c.language)
	params.Add("query", query)
	params.Add("page", "1")

	u := url.URL{
		Scheme: tmdbScheme,
		Host:   tmdbHost,
		Path:   path.Join(tmdbAPIv3, "search", "multi"), // ← /search/multi
	}

	requestURL := u.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = params.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("TMDB API request failed", "error", err, "query", query)
		return nil, fmt.Errorf("%w: %v", ErrAPIError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.logger.Error("Invalid TMDB API key")
		return nil, ErrInvalidAPIKey
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("TMDB API error", "status", resp.StatusCode, "query", query)
		return nil, fmt.Errorf("%w: status %d", ErrAPIError, resp.StatusCode)
	}

	var searchResp MultiSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		c.logger.Error("Failed to decode TMDB response", "error", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Фільтруємо персон (залишаємо тільки movies і tv)
	filtered := make([]MultiSearchResult, 0, len(searchResp.Results))
	for _, result := range searchResp.Results {
		if !result.IsPerson() {
			filtered = append(filtered, result)
		}
	}

	if len(filtered) == 0 {
		c.logger.Debug("Nothing found in TMDB", "query", query)
		return nil, ErrMovieNotFound
	}

	c.logger.Info("Results found in TMDB",
		"query", query,
		"count", len(filtered))

	return filtered, nil
}

// GetPosterURL повертає повний URL постера
// size: w92, w154, w185, w342, w500, w780, original
func (c *Client) GetPosterURL(posterPath string, size string) string {
	if posterPath == "" {
		return ""
	}

	u := url.URL{
		Scheme: posterScheme,
		Host:   posterHost,
		Path:   path.Join(posterBase, size, posterPath),
	}

	return u.String()
}
