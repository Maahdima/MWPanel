package model

type Peer struct {
	Model
	ServerID uint `gorm:"not null;default:0;index;uniqueIndex:uidx_peer_router_id;uniqueIndex:uidx_peer_router_address"`
	// A nil owner is explicitly superadmin-owned, including legacy/imported peers.
	ResellerID          *uint   `gorm:"index"`
	UUID                string  `gorm:"type:varchar(36);uniqueIndex;not null"`
	PeerID              string  `gorm:"type:varchar(255);uniqueIndex:uidx_peer_router_id;not null"`
	Disabled            bool    `gorm:"type:boolean;not null;default:false"`
	Comment             *string `gorm:"type:text"`
	Name                string  `gorm:"type:varchar(255);not null"`
	PrivateKey          string  `gorm:"type:varchar(255);not null"`
	PublicKey           string  `gorm:"type:varchar(255);not null"`
	Interface           string  `gorm:"type:varchar(255);not null"`
	AllowedAddress      string  `gorm:"type:varchar(255);uniqueIndex:uidx_peer_router_address;not null"`
	Endpoint            string  `gorm:"type:varchar(255);not null"`
	EndpointPort        string  `gorm:"type:varchar(10);not null"`
	PersistentKeepalive string  `gorm:"type:varchar(10)"`
	SchedulerID         *string `gorm:"type:varchar(255)"`
	QueueID             *string `gorm:"type:varchar(255)"`
	ExpireTime          *string `gorm:"type:varchar(255)"`
	TrafficLimit        *int64  `gorm:"type:bigint"`
	TelegramUsername    *string `gorm:"type:varchar(255)"`
	TrafficFirstNotify  bool    `gorm:"type:boolean;not null;default:false;column:traffic_first_notify"`  // 80%
	TrafficSecondNotify bool    `gorm:"type:boolean;not null;default:false;column:traffic_second_notify"` // 90%
	TrafficThirdNotify  bool    `gorm:"type:boolean;not null;default:false;column:traffic_third_notify"`  // 100%
	ExpireFirstNotify   bool    `gorm:"type:boolean;not null;default:false;column:expire_first_notify"`   // 3 days before
	ExpireSecondNotify  bool    `gorm:"type:boolean;not null;default:false;column:expire_second_notify"`  // 2 days before
	ExpireThirdNotify   bool    `gorm:"type:boolean;not null;default:false;column:expire_third_notify"`   // 1 day before
	DownloadBandwidth   *string `gorm:"type:varchar(255)"`
	UploadBandwidth     *string `gorm:"type:varchar(255)"`
	DownloadUsage       int64   `gorm:"type:bigint;not null;default:0"` // in bytes
	UploadUsage         int64   `gorm:"type:bigint;not null;default:0"` // in bytes
	LastTx              int64   `gorm:"type:bigint;not null;default:0"` // in bytes
	LastRx              int64   `gorm:"type:bigint;not null;default:0"` // in bytes
	IsShared            bool    `gorm:"type:boolean;not null;default:false"`
	ShareExpireTime     *string `gorm:"type:varchar(255)"`
}
