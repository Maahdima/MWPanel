package schema

type TelegramStatusResponse struct {
	Enabled     bool   `json:"enabled"`
	BotUsername string `json:"bot_username,omitempty"`
	BotURL      string `json:"bot_url,omitempty"`
}
