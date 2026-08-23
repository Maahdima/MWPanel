package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maahdima/mwp/api/config"

	"go.uber.org/zap"
)

const (
	telegramLongPollTimeout = 25
	telegramHTTPTimeout     = 40 * time.Second
)

type TelegramClient struct {
	botToken   string
	apiBaseURL string
	client     *http.Client
	logger     *zap.Logger
}

type telegramAPIResponse struct {
	Ok          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

type telegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type telegramChat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Username string `json:"username"`
}

type telegramMessage struct {
	MessageID int64         `json:"message_id"`
	From      *telegramUser `json:"from"`
	Chat      telegramChat  `json:"chat"`
	Text      string        `json:"text"`
}

type telegramCallbackQuery struct {
	ID      string           `json:"id"`
	From    telegramUser     `json:"from"`
	Message *telegramMessage `json:"message"`
	Data    string           `json:"data"`
}

type telegramUpdate struct {
	UpdateID      int64                  `json:"update_id"`
	Message       *telegramMessage       `json:"message"`
	CallbackQuery *telegramCallbackQuery `json:"callback_query"`
}

type telegramInlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

type telegramInlineKeyboardMarkup struct {
	InlineKeyboard [][]telegramInlineKeyboardButton `json:"inline_keyboard"`
}

type telegramSendMessageRequest struct {
	ChatID                string                        `json:"chat_id"`
	Text                  string                        `json:"text"`
	DisableWebPagePreview bool                          `json:"disable_web_page_preview"`
	ReplyMarkup           *telegramInlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type telegramGetUpdatesRequest struct {
	Offset         int64    `json:"offset"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates"`
}

type telegramBotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type telegramAnswerCallbackRequest struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
	ShowAlert       bool   `json:"show_alert,omitempty"`
}

func NewTelegramClient(cfg config.TelegramConfig) *TelegramClient {
	return &TelegramClient{
		botToken:   cfg.BotToken,
		apiBaseURL: strings.TrimRight(cfg.ApiBaseURL, "/"),
		client: &http.Client{
			Timeout: telegramHTTPTimeout,
		},
		logger: zap.L().Named("TelegramClient"),
	}
}

func (c *TelegramClient) GetMe(ctx context.Context) (*telegramUser, error) {
	var user telegramUser
	if err := c.call(ctx, "getMe", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *TelegramClient) DeleteWebhook(ctx context.Context) error {
	return c.call(ctx, "deleteWebhook", map[string]bool{"drop_pending_updates": false}, nil)
}

func (c *TelegramClient) SetMyCommands(ctx context.Context, commands []telegramBotCommand) error {
	return c.call(ctx, "setMyCommands", map[string]any{"commands": commands}, nil)
}

func (c *TelegramClient) GetUpdates(ctx context.Context, offset int64) ([]telegramUpdate, error) {
	var updates []telegramUpdate
	err := c.call(ctx, "getUpdates", telegramGetUpdatesRequest{
		Offset:         offset,
		Timeout:        telegramLongPollTimeout,
		AllowedUpdates: []string{"message", "callback_query"},
	}, &updates)
	if err != nil {
		return nil, err
	}
	return updates, nil
}

func (c *TelegramClient) SendMessage(ctx context.Context, chatID, text string, markup *telegramInlineKeyboardMarkup) error {
	return c.call(ctx, "sendMessage", telegramSendMessageRequest{
		ChatID:                chatID,
		Text:                  text,
		DisableWebPagePreview: true,
		ReplyMarkup:           markup,
	}, nil)
}

func (c *TelegramClient) AnswerCallbackQuery(ctx context.Context, callbackID, text string) error {
	return c.call(ctx, "answerCallbackQuery", telegramAnswerCallbackRequest{
		CallbackQueryID: callbackID,
		Text:            text,
	}, nil)
}

func (c *TelegramClient) SendDocument(ctx context.Context, chatID, filePath, filename string) error {
	return c.sendFile(ctx, "sendDocument", "document", chatID, filePath, filename, "")
}

func (c *TelegramClient) SendPhoto(ctx context.Context, chatID, filePath, caption string) error {
	return c.sendFile(ctx, "sendPhoto", "photo", chatID, filePath, filepath.Base(filePath), caption)
}

func (c *TelegramClient) sendFile(ctx context.Context, method, field, chatID, filePath, filename, caption string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return err
	}
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
	}

	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/bot%s/%s", c.apiBaseURL, c.botToken, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Telegram file send failed", zap.String("method", method), zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var parsed telegramAPIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("telegram %s failed with status %d", method, resp.StatusCode)
		}
		return err
	}
	if !parsed.Ok {
		return fmt.Errorf("telegram %s failed: %s", method, parsed.Description)
	}

	return nil
}

func (c *TelegramClient) call(ctx context.Context, method string, payload any, dest any) error {
	url := fmt.Sprintf("%s/bot%s/%s", c.apiBaseURL, c.botToken, method)

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}

	httpMethod := http.MethodGet
	if payload != nil {
		httpMethod = http.MethodPost
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, url, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("Telegram API request failed", zap.String("method", method), zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var parsed telegramAPIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("telegram %s failed with status %d", method, resp.StatusCode)
		}
		return err
	}
	if !parsed.Ok {
		return fmt.Errorf("telegram %s failed: %s", method, parsed.Description)
	}
	if dest == nil || len(parsed.Result) == 0 {
		return nil
	}

	return json.Unmarshal(parsed.Result, dest)
}
