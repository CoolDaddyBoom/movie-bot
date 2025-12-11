package telegram

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

type Client struct {
	host       string
	basePath   string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(token string, logger *slog.Logger) *Client {
	return &Client{
		host:       "api.telegram.org",
		basePath:   "bot" + token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
	}
}

func (c *Client) GetUpdates(offset int, limit int) ([]Update, error) {
	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(http.MethodGet, "getUpdates", q)
	if err != nil {
		return nil, err
	}

	var response UpdatesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (c *Client) SendMessage(chatID int, text string) error {
	c.logger.Debug("Sending message", "chat_id", chatID)
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)
	q.Add("parse_mode", "HTML")

	_, err := c.doRequest(http.MethodGet, "sendMessage", q)
	c.logger.Info("Message sent", "chat_id", chatID)
	return err
}

func (c *Client) SendPhoto(chatID int, photoURL string, caption string) error {
	c.logger.Debug("Sending photo", "chat_id", chatID, "photo_url", photoURL)

	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("photo", photoURL)
	q.Add("caption", caption)
	q.Add("parse_mode", "HTML") // підтримка HTML форматування

	_, err := c.doRequest(http.MethodGet, "sendPhoto", q) // можна POST, але GET теж має працювати
	c.logger.Info("Photo sent", "chat_id", chatID)
	return err
}

func (c *Client) SendPhotoOrMessage(chatID int, photoURL string, text string) error {
	if photoURL != "" {
		return c.SendPhoto(chatID, photoURL, text)
	} else {
		return c.SendMessage(chatID, text)
	}
}

func (c *Client) doRequest(httpMethod, apiMethod string, query url.Values) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, apiMethod),
	}

	url := u.String() // .String() конвертує структуру URL в рядок (з говна в те, що треба)

	req, err := http.NewRequest(httpMethod, url, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

/*1. Структура Client:
gotype Client struct {
    host       string        // "api.telegram.org"
    basePath   string        // "bot<TOKEN>"
    httpClient *http.Client  // для HTTP запитів з timeout
}
2. Конструктор NewClient:
gofunc NewClient(token string) *Client

Приймає токен
Будує basePath = "bot" + token
Створює httpClient з timeout 30 сек

3. Універсальний doRequest:
gofunc (c *Client) doRequest(httpMethod, apiMethod string, query url.Values) ([]byte, error)

Будує URL через url.URL структуру
Створює request через http.NewRequest
Додає query параметри через req.URL.RawQuery
Відправляє через client.Do()
Повертає []byte (raw body)

4. Публічні методи:
goGetUpdates(offset, limit int) ([]Update, error)
SendMessage(chatID int64, text string) error
```
- Створюють url.Values з параметрами
- Викликають doRequest
- Парсять JSON (GetUpdates) або просто повертають помилку (SendMessage)

**Ключова ідея:** doRequest - це DRY (Don't Repeat Yourself), вся HTTP логіка в одному місці.

---*/
