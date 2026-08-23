package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var errNoTelegramSubscribers = errors.New("no telegram chats are subscribed to alerts for this peer")

type TelegramNotifier struct {
	enabled bool
	db      *gorm.DB
	client  *TelegramClient
	logger  *zap.Logger
}

func NewTelegramNotifier(db *gorm.DB, cfg config.TelegramConfig, client *TelegramClient) *TelegramNotifier {
	return &TelegramNotifier{
		enabled: isTelegramEnabled(cfg),
		db:      db,
		client:  client,
		logger:  zap.L().Named("TelegramNotifier"),
	}
}

func (t *TelegramNotifier) Enabled() bool {
	return t.enabled
}

func (t *TelegramNotifier) NotifyPeerUsage(ctx context.Context, peer *model.Peer, percent int64, totalUsage, limit int64) error {
	if !t.enabled {
		return errors.New("telegram bot is disabled")
	}
	if t.client == nil {
		return errors.New("telegram client is not configured")
	}
	if peer == nil || peer.UUID == "" {
		return errors.New("peer is required")
	}

	var chats []model.TelegramChat
	if err := t.db.Where("peer_uuid = ?", peer.UUID).Find(&chats).Error; err != nil {
		return err
	}

	heading := "⚠️ Traffic alert"
	if percent >= 100 {
		heading = "🚨 Traffic limit reached"
	} else if percent >= 90 {
		heading = "🔥 High usage alert"
	}

	message := telegramBlocks(
		heading,
		"📡 Config: "+peer.Name,
		fmt.Sprintf("📊 Used: %d%%", percent),
		fmt.Sprintf("💾 %s GB of %s GB", utils.BytesToGB(totalUsage), utils.BytesToGB(limit)),
	)
	markup := t.alertMarkup(peer)

	var lastErr error
	sent := 0
	for _, chat := range chats {
		if !chat.NotifyEnabled || chat.ChatID == "" {
			continue
		}
		if err := t.client.SendMessage(ctx, chat.ChatID, message, markup); err != nil {
			t.logger.Error("Telegram send failed", zap.String("chatID", chat.ChatID), zap.Error(err))
			lastErr = err
			continue
		}
		sent++
	}

	if sent == 0 {
		if lastErr != nil {
			return lastErr
		}
		return errNoTelegramSubscribers
	}

	return nil
}

func (t *TelegramNotifier) alertMarkup(peer *model.Peer) *telegramInlineKeyboardMarkup {
	if peer == nil || peer.UUID == "" {
		return nil
	}

	return &telegramInlineKeyboardMarkup{
		InlineKeyboard: [][]telegramInlineKeyboardButton{
			{{Text: "📊 Details", CallbackData: "details:" + peer.UUID}},
		},
	}
}
