package service

import (
	"strings"

	"github.com/google/uuid"

	"github.com/maahdima/mwp/api/config"
)

func isTelegramEnabled(cfg config.TelegramConfig) bool {
	return cfg.Enabled && strings.TrimSpace(cfg.BotToken) != "" && strings.TrimSpace(cfg.ApiBaseURL) != ""
}

func telegramBlocks(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, "\n\n")
}

func NormalizeTelegramUsername(username string) string {
	normalized := strings.TrimSpace(username)
	normalized = strings.TrimPrefix(normalized, "@")
	return strings.ToLower(normalized)
}

func NormalizeTelegramUsernamePtr(username *string) *string {
	if username == nil {
		return nil
	}

	normalized := NormalizeTelegramUsername(*username)
	if normalized == "" {
		return nil
	}

	return &normalized
}

func parseBotCommand(text string) (command string, payload string) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return "", ""
	}

	parts := strings.SplitN(trimmed, " ", 2)
	command = strings.ToLower(parts[0])
	if at := strings.Index(command, "@"); at >= 0 {
		command = command[:at]
	}
	command = strings.TrimPrefix(command, "/")

	if len(parts) > 1 {
		payload = strings.TrimSpace(parts[1])
	}

	return command, payload
}

func parseCallbackData(data string) (action, peerUUID string) {
	parts := strings.SplitN(strings.TrimSpace(data), ":", 2)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func parsePeerUUID(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", false
	}
	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
}
