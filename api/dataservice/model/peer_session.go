package model

type PeerSession struct {
	Model
	PeerID         uint    `gorm:"index:idx_peer_sessions_peer;not null"`
	PeerUUID       string  `gorm:"type:varchar(36);index;not null"`
	ConnectedAt    int64   `gorm:"index;not null"`
	DisconnectedAt *int64  `gorm:"index:idx_peer_sessions_open"`
	Endpoint       *string `gorm:"type:varchar(255)"`
	EndpointPort   *string `gorm:"type:varchar(16)"`
	StartDownload  int64   `gorm:"not null;default:0"`
	StartUpload    int64   `gorm:"not null;default:0"`
	DownloadUsage  int64   `gorm:"not null;default:0"`
	UploadUsage    int64   `gorm:"not null;default:0"`
}
