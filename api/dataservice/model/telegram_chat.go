package model

type TelegramChat struct {
	Model
	ChatID        string  `gorm:"type:varchar(64);not null;uniqueIndex:idx_telegram_chat_peer"`
	PeerUUID      string  `gorm:"type:varchar(36);not null;uniqueIndex:idx_telegram_chat_peer;index"`
	Username      *string `gorm:"type:varchar(255)"`
	NotifyEnabled bool    `gorm:"type:boolean;not null"`
}
