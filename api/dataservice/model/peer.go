package model

type Peer struct {
	Model
	UUID                string  `gorm:"type:varchar(36);uniqueIndex;not null"`
	PeerID              string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Disabled            bool    `gorm:"type:boolean;not null;default:false"`
	Comment             *string `gorm:"type:text"`
	Name                string  `gorm:"type:varchar(255);not null"`
	PrivateKey          string  `gorm:"type:varchar(255);not null"`
	PublicKey           string  `gorm:"type:varchar(255);not null"`
	Interface           string  `gorm:"type:varchar(255);not null"`
	AllowedAddress      string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Endpoint            string  `gorm:"type:varchar(255);not null"`
	EndpointPort        string  `gorm:"type:varchar(10);not null"`
	PersistentKeepalive string  `gorm:"type:varchar(10)"`
	SchedulerID         *string `gorm:"type:varchar(255)"`
	QueueID             *string `gorm:"type:varchar(255)"`
	ExpireTime          *string `gorm:"type:varchar(255)"`
	TrafficLimit        *int64  `gorm:"type:bigint;default:0"` // in bytes, 0 means no limit
	DownloadBandwidth   *string `gorm:"type:varchar(255)"`
	UploadBandwidth     *string `gorm:"type:varchar(255)"`
	DownloadUsage       int64   `gorm:"type:bigint;not null;default:0"` // in bytes
	UploadUsage         int64   `gorm:"type:bigint;not null;default:0"` // in bytes
	LastTx              int64   `gorm:"type:bigint;not null;default:0"` // in bytes
	LastRx              int64   `gorm:"type:bigint;not null;default:0"` // in bytes
	IsShared            bool    `gorm:"type:boolean;not null;default:false"`
	ShareExpireTime     *string `gorm:"type:varchar(255)"`
}
