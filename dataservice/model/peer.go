package model

import "gorm.io/gorm"

type Peer struct {
	gorm.Model
	PeerID              string  `gorm:"uniqueIndex;not null"`
	Disabled            bool    `gorm:"column:disabled"`
	Comment             *string `gorm:""`
	PeerName            string  `gorm:""`
	PublicKey           string  `gorm:"not null"`
	Interface           string  `gorm:"not null"`
	AllowedAddress      string  `gorm:""`
	Endpoint            string  `gorm:"not null"`
	EndpointPort        string  `gorm:"not null"`
	PersistentKeepalive string  `gorm:""`
	SchedulerID         *string `gorm:""`
	QueueID             *string `gorm:""`
	ExpireTime          *string `gorm:""`
	TrafficLimit        *string `gorm:""`
	DownloadBandwidth   *string `gorm:""`
	UploadBandwidth     *string `gorm:""`
	DownloadUsage       string  `gorm:"default:0"`
	UploadUsage         string  `gorm:"default:0"`
}
