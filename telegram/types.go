package telegram

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	UpdateID int              `json:"update_id"`
	Message  *IncomingMessage `json:"message"`
}

type IncomingMessage struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	From From   `json:"from"`
}

type Chat struct {
	ID int `json:"id"`
}

type From struct {
	Username string `json:"username"`
}
