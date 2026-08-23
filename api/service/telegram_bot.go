package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"
)

type botSessionState int

const (
	sessionIdle botSessionState = iota
	sessionAwaitingUUID
	sessionAwaitingUsername
)

type botSession struct {
	state    botSessionState
	peerUUID string
}

type TelegramStatus struct {
	Enabled     bool
	BotUsername string
	BotURL      string
}

type TelegramBot struct {
	enabled         bool
	db              *gorm.DB
	client          *TelegramClient
	configGenerator *ConfigGenerator
	qrCodeGenerator *QRCodeGenerator
	mikrotikAdaptor *mikrotik.Adaptor
	logger          *zap.Logger
	mu              sync.RWMutex
	botUsername     string
	sessions        map[string]*botSession
	sessionsMu      sync.Mutex
}

func NewTelegramBot(
	db *gorm.DB,
	cfg config.TelegramConfig,
	client *TelegramClient,
	configGenerator *ConfigGenerator,
	qrCodeGenerator *QRCodeGenerator,
	mikrotikAdaptor *mikrotik.Adaptor,
) *TelegramBot {
	return &TelegramBot{
		enabled:         isTelegramEnabled(cfg),
		db:              db,
		client:          client,
		configGenerator: configGenerator,
		qrCodeGenerator: qrCodeGenerator,
		mikrotikAdaptor: mikrotikAdaptor,
		logger:          zap.L().Named("TelegramBot"),
		sessions:        make(map[string]*botSession),
	}
}

func (b *TelegramBot) Start(ctx context.Context) {
	if !b.enabled || b.client == nil {
		b.logger.Info("Telegram bot is disabled")
		return
	}

	go b.run(ctx)
}

func (b *TelegramBot) Status() TelegramStatus {
	if !b.enabled {
		return TelegramStatus{Enabled: false}
	}

	username := b.cachedBotUsername()
	if username == "" && b.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if me, err := b.client.GetMe(ctx); err == nil {
			b.setBotUsername(me.Username)
			username = me.Username
		}
	}

	status := TelegramStatus{
		Enabled:     true,
		BotUsername: username,
	}
	if username != "" {
		status.BotURL = "https://t.me/" + username
	}

	return status
}

func (b *TelegramBot) run(ctx context.Context) {
	if err := b.client.DeleteWebhook(ctx); err != nil {
		b.logger.Warn("Failed to delete Telegram webhook before polling", zap.Error(err))
	}

	if err := b.client.SetMyCommands(ctx, []telegramBotCommand{
		{Command: "start", Description: "🔗 Link your config"},
		{Command: "menu", Description: "📋 Open the action menu"},
		{Command: "details", Description: "📊 Usage, status and expiry"},
		{Command: "config", Description: "📄 Download WireGuard config"},
		{Command: "qrcode", Description: "📱 Show config QR code"},
		{Command: "notify", Description: "🔔 Turn traffic alerts on/off"},
		{Command: "unlink", Description: "🔓 Unlink this chat"},
		{Command: "help", Description: "❓ How to use the bot"},
	}); err != nil {
		b.logger.Warn("Failed to set Telegram bot commands", zap.Error(err))
	}

	if me, err := b.client.GetMe(ctx); err != nil {
		b.logger.Error("Failed to fetch Telegram bot identity", zap.Error(err))
	} else {
		b.setBotUsername(me.Username)
		b.logger.Info("Telegram bot started", zap.String("username", me.Username))
	}

	offset := int64(0)
	for {
		if ctx.Err() != nil {
			return
		}

		updates, err := b.client.GetUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			b.logger.Error("Failed to poll Telegram updates", zap.Error(err))
			time.Sleep(3 * time.Second)
			continue
		}

		for _, update := range updates {
			b.handleUpdate(ctx, update)
			offset = update.UpdateID + 1
		}
	}
}

func (b *TelegramBot) handleUpdate(ctx context.Context, update telegramUpdate) {
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, update.CallbackQuery)
		return
	}
	if update.Message == nil {
		return
	}

	msg := update.Message
	if msg.Chat.Type != "" && msg.Chat.Type != "private" {
		return
	}

	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	command, payload := parseBotCommand(msg.Text)
	switch command {
	case "start":
		b.handleStart(ctx, msg, payload)
	case "help":
		b.reply(ctx, chatID, b.helpText(), nil)
	case "menu", "details", "config", "qrcode", "notify", "unlink":
		b.handleNamedCommand(ctx, msg, command, payload)
	case "skip":
		b.handleSkipUsername(ctx, chatID)
	default:
		b.handleText(ctx, msg)
	}
}

func (b *TelegramBot) handleStart(ctx context.Context, msg *telegramMessage, payload string) {
	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	if peerUUID, ok := parsePeerUUID(payload); ok {
		b.linkPeer(ctx, msg, peerUUID)
		return
	}

	b.setSession(chatID, &botSession{state: sessionAwaitingUUID})
	b.reply(ctx, chatID, telegramBlocks(
		"👋 Welcome to MWPanel!",
		"🔗 Send your config UUID to link this chat.",
		"💡 Tip: open the “Connect to Telegram bot” button on your share page — it already includes the UUID.",
	), nil)
}

func (b *TelegramBot) handleText(ctx context.Context, msg *telegramMessage) {
	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	text := strings.TrimSpace(msg.Text)
	session := b.getSession(chatID)

	if peerUUID, ok := parsePeerUUID(text); ok {
		b.linkPeer(ctx, msg, peerUUID)
		return
	}

	if session != nil && session.state == sessionAwaitingUsername {
		username := NormalizeTelegramUsername(text)
		if username == "" || username == "skip" {
			b.handleSkipUsername(ctx, chatID)
			return
		}
		if err := b.saveUsername(chatID, session.peerUUID, username); err != nil {
			b.logger.Error("Failed to save Telegram username", zap.Error(err))
			b.reply(ctx, chatID, "⚠️ Could not save that username. Try again, or send /skip.", nil)
			return
		}
		b.setSession(chatID, nil)
		b.sendPeerMenu(ctx, chatID, session.peerUUID, fmt.Sprintf("✅ Saved username @%s.", username))
		return
	}

	if session != nil && session.state == sessionAwaitingUUID {
		b.reply(ctx, chatID, "🤔 That doesn't look like a config UUID.\n\nPaste the UUID from the panel and send it here.", nil)
		return
	}

	if strings.HasPrefix(text, "/") {
		b.reply(ctx, chatID, b.helpText(), nil)
		return
	}

	b.reply(ctx, chatID, telegramBlocks(
		"👋 Send /start and then your config UUID to link this chat.",
		"❓ Send /help for the list of commands.",
	), nil)
}

func (b *TelegramBot) handleNamedCommand(ctx context.Context, msg *telegramMessage, command, payload string) {
	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	peerUUID, err := b.resolvePeerUUID(chatID, payload)
	if err != nil {
		b.reply(ctx, chatID, err.Error(), nil)
		return
	}

	switch command {
	case "menu":
		b.sendPeerMenu(ctx, chatID, peerUUID, "")
	case "details":
		b.sendDetails(ctx, chatID, peerUUID)
	case "config":
		b.sendConfig(ctx, chatID, peerUUID)
	case "qrcode":
		b.sendQRCode(ctx, chatID, peerUUID)
	case "notify":
		b.toggleNotify(ctx, chatID, peerUUID)
	case "unlink":
		b.unlinkPeer(ctx, chatID, peerUUID)
	}
}

func (b *TelegramBot) handleCallback(ctx context.Context, query *telegramCallbackQuery) {
	if query.Message == nil && query.From.ID == 0 {
		return
	}

	chatID := strconv.FormatInt(query.From.ID, 10)
	if query.Message != nil {
		chatID = strconv.FormatInt(query.Message.Chat.ID, 10)
	}
	action, peerUUID := parseCallbackData(query.Data)
	ack := ""

	switch action {
	case "menu":
		b.sendPeerMenu(ctx, chatID, peerUUID, "")
	case "details":
		b.sendDetails(ctx, chatID, peerUUID)
	case "config":
		b.sendConfig(ctx, chatID, peerUUID)
		ack = "📄 Sending config…"
	case "qrcode":
		b.sendQRCode(ctx, chatID, peerUUID)
		ack = "📱 Sending QR code…"
	case "notify":
		b.toggleNotify(ctx, chatID, peerUUID)
	case "skipuser":
		b.handleSkipUsername(ctx, chatID)
		ack = "⏭️ Skipped username"
	case "unlink":
		b.unlinkPeer(ctx, chatID, peerUUID)
	default:
		ack = "Unknown action"
	}

	if err := b.client.AnswerCallbackQuery(ctx, query.ID, ack); err != nil {
		b.logger.Warn("Failed to answer Telegram callback", zap.Error(err))
	}
}

func (b *TelegramBot) handleSkipUsername(ctx context.Context, chatID string) {
	session := b.getSession(chatID)
	peerUUID := ""
	if session != nil {
		peerUUID = session.peerUUID
	}
	b.setSession(chatID, nil)
	if peerUUID != "" {
		b.sendPeerMenu(ctx, chatID, peerUUID, "✅ No problem — your chat is already linked, so alerts can still be delivered.")
		return
	}
	b.reply(ctx, chatID, "✅ No problem — your chat is already linked, so alerts can still be delivered.", nil)
}

func (b *TelegramBot) linkPeer(ctx context.Context, msg *telegramMessage, peerUUID string) {
	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	username := messageUsername(msg)

	peer, err := b.findPeer(peerUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			b.reply(ctx, chatID, "❌ No config found for that UUID.\n\nCheck the value from the panel and try again.", nil)
			return
		}
		b.logger.Error("Failed to look up peer for Telegram link", zap.Error(err), zap.String("uuid", peerUUID))
		b.reply(ctx, chatID, "⚠️ Could not link that config right now. Please try again in a moment.", nil)
		return
	}

	if err := b.upsertChat(chatID, peer.UUID, username); err != nil {
		b.logger.Error("Failed to store Telegram chat link", zap.Error(err), zap.String("uuid", peer.UUID))
		b.reply(ctx, chatID, "⚠️ Could not link that config right now. Please try again in a moment.", nil)
		return
	}

	if username == "" {
		b.setSession(chatID, &botSession{state: sessionAwaitingUsername, peerUUID: peer.UUID})
		skipMarkup := &telegramInlineKeyboardMarkup{
			InlineKeyboard: [][]telegramInlineKeyboardButton{
				{{Text: "⏭️ Skip", CallbackData: "skipuser:" + peer.UUID}},
			},
		}
		b.reply(ctx, chatID, telegramBlocks(
			"✅ Linked config: "+peer.Name,
			"👤 I don't see a Telegram username on your account.",
			"Send it now (with or without @), or tap Skip.",
			"💬 Your chat is already saved — alerts will still work.",
		), skipMarkup)
		return
	}

	b.setSession(chatID, nil)
	b.sendPeerMenu(ctx, chatID, peer.UUID, fmt.Sprintf("✅ Linked as @%s", username))
}

func (b *TelegramBot) upsertChat(chatID, peerUUID, username string) error {
	if chatID == "" || peerUUID == "" {
		return errors.New("telegram chat ID and peer UUID are required")
	}

	username = NormalizeTelegramUsername(username)
	var usernamePtr *string
	if username != "" {
		usernamePtr = &username
	}

	var existing model.TelegramChat
	err := b.db.Where("chat_id = ? AND peer_uuid = ?", chatID, peerUUID).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{}
		if usernamePtr != nil {
			updates["username"] = username
		}
		if len(updates) == 0 {
			return nil
		}
		return b.db.Model(&existing).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return b.db.Create(&model.TelegramChat{
		ChatID:        chatID,
		PeerUUID:      peerUUID,
		Username:      usernamePtr,
		NotifyEnabled: true,
	}).Error
}

func (b *TelegramBot) saveUsername(chatID, peerUUID, username string) error {
	username = NormalizeTelegramUsername(username)
	if username == "" {
		return errors.New("username is empty")
	}
	return b.db.Model(&model.TelegramChat{}).
		Where("chat_id = ? AND peer_uuid = ?", chatID, peerUUID).
		Update("username", username).Error
}

func (b *TelegramBot) sendPeerMenu(ctx context.Context, chatID, peerUUID, intro string) {
	chat, peer, err := b.linkedPeer(chatID, peerUUID)
	if err != nil {
		b.reply(ctx, chatID, "🔒 This chat is not linked to that config.\n\nSend /start and the config UUID first.", nil)
		return
	}

	notifyLabel := "🔔 Alerts: ON"
	if !chat.NotifyEnabled {
		notifyLabel = "🔕 Alerts: OFF"
	}

	markup := &telegramInlineKeyboardMarkup{
		InlineKeyboard: [][]telegramInlineKeyboardButton{
			{{Text: "📊 Details", CallbackData: "details:" + peer.UUID}},
			{
				{Text: "📄 Download config", CallbackData: "config:" + peer.UUID},
				{Text: "📱 QR code", CallbackData: "qrcode:" + peer.UUID},
			},
			{{Text: notifyLabel, CallbackData: "notify:" + peer.UUID}},
			{{Text: "🔓 Unlink", CallbackData: "unlink:" + peer.UUID}},
		},
	}

	text := telegramBlocks(intro, "📡 Config: "+peer.Name, "👇 Choose an action:")
	b.reply(ctx, chatID, text, markup)
}

func (b *TelegramBot) sendDetails(ctx context.Context, chatID, peerUUID string) {
	chat, peer, err := b.linkedPeer(chatID, peerUUID)
	if err != nil {
		b.reply(ctx, chatID, "🔒 This chat is not linked to that config.\n\nSend /start and the config UUID first.", nil)
		return
	}

	b.reply(ctx, chatID, b.formatPeerDetails(peer, chat), b.peerMenuMarkup(peer, chat.NotifyEnabled))
}

func (b *TelegramBot) sendConfig(ctx context.Context, chatID, peerUUID string) {
	_, peer, err := b.linkedPeer(chatID, peerUUID)
	if err != nil {
		b.reply(ctx, chatID, "🔒 This chat is not linked to that config.\n\nSend /start and the config UUID first.", nil)
		return
	}
	if b.configGenerator == nil {
		b.reply(ctx, chatID, "⚠️ Config download is not available right now.", nil)
		return
	}

	path, err := b.configGenerator.GetConfigByUUID(peer.UUID)
	if err != nil {
		b.reply(ctx, chatID, "❌ Could not find that config file.", nil)
		return
	}

	filename := peer.Name + ".conf"
	if err := b.client.SendDocument(ctx, chatID, path, filename); err != nil {
		b.logger.Error("Failed to send Telegram config", zap.Error(err), zap.String("uuid", peer.UUID))
		b.reply(ctx, chatID, "⚠️ Could not send the config file right now.", nil)
	}
}

func (b *TelegramBot) sendQRCode(ctx context.Context, chatID, peerUUID string) {
	_, peer, err := b.linkedPeer(chatID, peerUUID)
	if err != nil {
		b.reply(ctx, chatID, "🔒 This chat is not linked to that config.\n\nSend /start and the config UUID first.", nil)
		return
	}
	if b.qrCodeGenerator == nil {
		b.reply(ctx, chatID, "⚠️ QR code is not available right now.", nil)
		return
	}

	path, err := b.qrCodeGenerator.GetQRCodeByUUID(peer.UUID)
	if err != nil {
		b.reply(ctx, chatID, "❌ Could not find that QR code.", nil)
		return
	}

	if err := b.client.SendPhoto(ctx, chatID, path, "📱 QR code for "+peer.Name); err != nil {
		b.logger.Error("Failed to send Telegram QR code", zap.Error(err), zap.String("uuid", peer.UUID))
		b.reply(ctx, chatID, "⚠️ Could not send the QR code right now.", nil)
	}
}

func (b *TelegramBot) toggleNotify(ctx context.Context, chatID, peerUUID string) {
	chat, peer, err := b.linkedPeer(chatID, peerUUID)
	if err != nil {
		b.reply(ctx, chatID, "🔒 This chat is not linked to that config.\n\nSend /start and the config UUID first.", nil)
		return
	}

	enabled := !chat.NotifyEnabled
	if err := b.db.Model(&chat).Update("notify_enabled", enabled).Error; err != nil {
		b.logger.Error("Failed to toggle Telegram notifications", zap.Error(err))
		b.reply(ctx, chatID, "⚠️ Could not update notification settings.", nil)
		return
	}

	message := "🔕 Traffic alerts are now OFF for " + peer.Name + "."
	if enabled {
		message = "🔔 Traffic alerts are now ON for " + peer.Name + "."
	}
	b.reply(ctx, chatID, message, b.peerMenuMarkup(peer, enabled))
}

func (b *TelegramBot) unlinkPeer(ctx context.Context, chatID, peerUUID string) {
	_, peer, err := b.linkedPeer(chatID, peerUUID)
	if err != nil {
		b.reply(ctx, chatID, "🔒 This chat is not linked to that config.", nil)
		return
	}

	if err := b.db.Unscoped().Where("chat_id = ? AND peer_uuid = ?", chatID, peerUUID).Delete(&model.TelegramChat{}).Error; err != nil {
		b.logger.Error("Failed to unlink Telegram chat", zap.Error(err))
		b.reply(ctx, chatID, "⚠️ Could not unlink this chat right now.", nil)
		return
	}

	b.setSession(chatID, nil)
	b.reply(ctx, chatID, telegramBlocks(
		"🔓 Unlinked config "+peer.Name+".",
		"Send /start with a UUID to link again.",
	), nil)
}

func (b *TelegramBot) formatPeerDetails(peer model.Peer, chat model.TelegramChat) string {
	totalUsage := peer.DownloadUsage + peer.UploadUsage
	status := "🔴 Offline"
	if b.isPeerOnline(peer) {
		status = "🟢 Online"
	}

	traffic := "♾️ Unlimited"
	if peer.TrafficLimit != nil && *peer.TrafficLimit > 0 {
		percent := float64(totalUsage) / float64(*peer.TrafficLimit) * 100
		traffic = fmt.Sprintf("%s / %s GB (%.1f%%)", utils.BytesToGB(totalUsage), utils.BytesToGB(*peer.TrafficLimit), percent)
	} else {
		traffic = utils.BytesToGB(totalUsage) + " GB used"
	}

	expire := "♾️ Never"
	if peer.ExpireTime != nil && strings.TrimSpace(*peer.ExpireTime) != "" {
		expire = *peer.ExpireTime
	}

	alerts := "On"
	if !chat.NotifyEnabled {
		alerts = "Off"
	}

	username := "—"
	if chat.Username != nil && *chat.Username != "" {
		username = "@" + *chat.Username
	}

	return telegramBlocks(
		"📡 Config: "+peer.Name,
		"📶 Status: "+status,
		"📊 Traffic: "+traffic,
		"⬇️ Download: "+utils.BytesToGB(peer.DownloadUsage)+" GB",
		"⬆️ Upload: "+utils.BytesToGB(peer.UploadUsage)+" GB",
		"⏳ Expire: "+expire,
		"🔔 Alerts: "+alerts,
		"👤 Username: "+username,
	)
}

func (b *TelegramBot) isPeerOnline(peer model.Peer) bool {
	if b.mikrotikAdaptor == nil {
		return false
	}

	mtPeer, err := b.mikrotikAdaptor.FetchWgPeer(context.Background(), peer.PeerID)
	if err != nil || mtPeer.LastHandshake == nil {
		return false
	}

	_, online, err := utils.HandshakeStatus(mtPeer.Disabled, mtPeer.LastHandshake, common.PeerOnlineHandshakeTimeout)
	return err == nil && online
}

func (b *TelegramBot) peerMenuMarkup(peer model.Peer, notifyEnabled bool) *telegramInlineKeyboardMarkup {
	notifyLabel := "🔔 Alerts: ON"
	if !notifyEnabled {
		notifyLabel = "🔕 Alerts: OFF"
	}
	return &telegramInlineKeyboardMarkup{
		InlineKeyboard: [][]telegramInlineKeyboardButton{
			{{Text: "📊 Details", CallbackData: "details:" + peer.UUID}},
			{
				{Text: "📄 Download config", CallbackData: "config:" + peer.UUID},
				{Text: "📱 QR code", CallbackData: "qrcode:" + peer.UUID},
			},
			{{Text: notifyLabel, CallbackData: "notify:" + peer.UUID}},
		},
	}
}

func (b *TelegramBot) resolvePeerUUID(chatID, payload string) (string, error) {
	if peerUUID, ok := parsePeerUUID(payload); ok {
		if _, _, err := b.linkedPeer(chatID, peerUUID); err != nil {
			return "", errors.New("🔒 This chat is not linked to that config.\n\nSend /start and the UUID first.")
		}
		return peerUUID, nil
	}

	var chats []model.TelegramChat
	if err := b.db.Where("chat_id = ?", chatID).Find(&chats).Error; err != nil {
		return "", errors.New("⚠️ Could not look up linked configs.")
	}
	if len(chats) == 0 {
		return "", errors.New("🔒 This chat is not linked yet.\n\nSend /start and then your config UUID.")
	}
	if len(chats) > 1 {
		return "", errors.New("📋 This chat is linked to multiple configs.\n\nOpen /menu from the last link message, or send /start with a specific UUID.")
	}
	return chats[0].PeerUUID, nil
}

func (b *TelegramBot) linkedPeer(chatID, peerUUID string) (model.TelegramChat, model.Peer, error) {
	var chat model.TelegramChat
	if err := b.db.Where("chat_id = ? AND peer_uuid = ?", chatID, peerUUID).First(&chat).Error; err != nil {
		return model.TelegramChat{}, model.Peer{}, err
	}

	peer, err := b.findPeer(peerUUID)
	if err != nil {
		return model.TelegramChat{}, model.Peer{}, err
	}

	return chat, peer, nil
}

func (b *TelegramBot) findPeer(peerUUID string) (model.Peer, error) {
	var peer model.Peer
	err := b.db.Where("uuid = ?", peerUUID).First(&peer).Error
	return peer, err
}

func (b *TelegramBot) helpText() string {
	return telegramBlocks(
		"🤖 MWPanel Telegram bot",
		"🔗 Link your config by sending /start and then the UUID from the panel.",
		"After that you can:",
		"📊 /details — usage, status and expiry",
		"📄 /config — download the WireGuard file",
		"📱 /qrcode — show the QR code",
		"🔔 /notify — turn traffic alerts on or off",
		"📋 /menu — show the buttons again",
		"🔓 /unlink — remove this chat from the config",
		"❓ /help — show this message",
	)
}

func (b *TelegramBot) reply(ctx context.Context, chatID, text string, markup *telegramInlineKeyboardMarkup) {
	if b.client == nil {
		return
	}
	if err := b.client.SendMessage(ctx, chatID, text, markup); err != nil {
		b.logger.Error("Failed to send Telegram reply", zap.Error(err))
	}
}

func (b *TelegramBot) getSession(chatID string) *botSession {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()
	return b.sessions[chatID]
}

func (b *TelegramBot) setSession(chatID string, session *botSession) {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()
	if session == nil {
		delete(b.sessions, chatID)
		return
	}
	b.sessions[chatID] = session
}

func (b *TelegramBot) cachedBotUsername() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.botUsername
}

func (b *TelegramBot) setBotUsername(username string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.botUsername = strings.TrimPrefix(strings.TrimSpace(username), "@")
}

func messageUsername(msg *telegramMessage) string {
	if msg.From != nil {
		if username := NormalizeTelegramUsername(msg.From.Username); username != "" {
			return username
		}
	}
	return NormalizeTelegramUsername(msg.Chat.Username)
}
