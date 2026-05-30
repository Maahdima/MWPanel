package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/utils"

	"go.uber.org/zap"
)

type TelegramNotifier struct {
	enabled    bool
	botToken   string
	apiBaseURL string
	client     *http.Client
	logger     *zap.Logger
}

type telegramSendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type telegramSendMessageResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

func NewTelegramNotifier(cfg config.TelegramConfig) *TelegramNotifier {
	apiBaseURL := strings.TrimRight(cfg.ApiBaseURL, "/")
	enabled := cfg.Enabled && cfg.BotToken != "" && apiBaseURL != ""

	return &TelegramNotifier{
		enabled:    enabled,
		botToken:   cfg.BotToken,
		apiBaseURL: apiBaseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: zap.L().Named("TelegramNotifier"),
	}
}

func (t *TelegramNotifier) NotifyPeerUsage(ctx context.Context, peerName, telegramUsername string, percent int64, totalUsage, limit int64) error {
	if !t.enabled {
		return nil
	}

	username := strings.TrimSpace(telegramUsername)
	if username == "" {
		return nil
	}
	if !strings.HasPrefix(username, "@") {
		username = "@" + username
	}

	message := fmt.Sprintf(
		"Traffic alert for %s: %d%% used (%s GB of %s GB).",
		peerName,
		percent,
		utils.BytesToGB(totalUsage),
		utils.BytesToGB(limit),
	)

	payload := telegramSendMessageRequest{
		ChatID:                username,
		Text:                  message,
		DisableWebPagePreview: true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", t.apiBaseURL, t.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		t.logger.Error("Telegram send failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		t.logger.Warn("Telegram send returned non-200", zap.Int("status", resp.StatusCode), zap.String("body", string(respBody)))
		return fmt.Errorf("telegram send failed with status %d", resp.StatusCode)
	}

	var parsed telegramSendMessageResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return err
	}
	if !parsed.Ok {
		return fmt.Errorf("telegram send failed: %s", parsed.Description)
	}

	return nil
}

