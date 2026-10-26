package model

// RouterMigration and its rows are a durable, resumable workflow, independent
// of accounting history. Source snapshots remain available until cleanup ends.
type RouterMigration struct {
	ID                     string `gorm:"type:varchar(36);primaryKey"`
	CreatedAt              int64  `gorm:"autoCreateTime"`
	ResellerID             uint   `gorm:"not null;uniqueIndex:uidx_router_migration_request"`
	IdempotencyKey         string `gorm:"type:varchar(128);not null;uniqueIndex:uidx_router_migration_request"`
	RequestHash            string `gorm:"type:varchar(64);not null"`
	DestinationInterfaceID uint   `gorm:"not null"`
	ActorID                uint   `gorm:"not null"`
	Reason                 string `gorm:"type:text;not null"`
	State                  string `gorm:"type:varchar(16);not null;index"`
	LeaseUntil             int64  `gorm:"not null;default:0"`
	NextAttemptAt          int64  `gorm:"not null;default:0;index"`
	Attempts               int64  `gorm:"not null;default:0"`
	LastError              string `gorm:"type:text"`
}

type RouterMigrationPeer struct {
	ID                  string `gorm:"type:varchar(36);primaryKey"`
	MigrationID         string `gorm:"type:varchar(36);not null;uniqueIndex:uidx_router_migration_peer"`
	PeerID              uint   `gorm:"not null;uniqueIndex:uidx_router_migration_peer"`
	PublicKey           string `gorm:"type:varchar(255);not null;index"`
	SourceServerID      uint   `gorm:"not null;index"`
	SourceInterfaceID   uint   `gorm:"not null"`
	SourcePeerID        string `gorm:"type:varchar(255);not null"`
	DestinationServerID uint   `gorm:"not null;index"`
	SourceSnapshot      string `gorm:"type:text;not null"`
	DestinationSnapshot string `gorm:"type:text;not null"`
	Phase               string `gorm:"type:varchar(16);not null;index"`
}
