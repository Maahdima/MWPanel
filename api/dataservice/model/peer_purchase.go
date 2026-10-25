package model

// PeerPurchase is a durable outbox. Its hold and peer lock are created in the
// same transaction. External calls never run inside the financial transaction.
type PeerPurchase struct {
	ID             string `gorm:"type:varchar(36);primaryKey"`
	CreatedAt      int64  `gorm:"autoCreateTime"`
	ResellerID     uint   `gorm:"not null;uniqueIndex:uidx_peer_purchase_request"`
	IdempotencyKey string `gorm:"type:varchar(128);not null;uniqueIndex:uidx_peer_purchase_request"`
	RequestHash    string `gorm:"type:varchar(64);not null"`
	ChargeID       *uint  `gorm:"uniqueIndex"`
	PeerID         uint   `gorm:"not null;index"`
	ActorID        uint   `gorm:"not null"`
	Payload        string `gorm:"type:text"`
	Kind           string `gorm:"type:varchar(32);not null"`
	State          string `gorm:"type:varchar(16);not null;index"`
	QuotaBytes     int64  `gorm:"not null"`
	ExpireTime     string `gorm:"not null"`
	ResetUsage     bool   `gorm:"not null"`
	CapturedTx     int64  `gorm:"not null"`
	CapturedRx     int64  `gorm:"not null"`
	LeaseUntil     int64  `gorm:"not null;default:0"`
	NextAttemptAt  int64  `gorm:"not null;default:0;index"`
	Attempts       int64  `gorm:"not null;default:0"`
	LastError      string `gorm:"type:text"`
}
