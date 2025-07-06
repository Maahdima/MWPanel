package model

import "gorm.io/gorm"

type Peer struct {
	gorm.Model
	PeerID              string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Disabled            bool    `gorm:"type:boolean;not null;default:false"`
	Comment             *string `gorm:"type:text"`
	PeerName            string  `gorm:"type:varchar(255);not null"`
	PublicKey           string  `gorm:"type:varchar(255);not null"`
	Interface           string  `gorm:"type:varchar(255);not null"`
	AllowedAddress      string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Endpoint            string  `gorm:"type:varchar(255);not null"`
	EndpointPort        string  `gorm:"type:varchar(10);not null"`
	PersistentKeepalive string  `gorm:"type:varchar(10)"`
	SchedulerID         *string `gorm:"type:varchar(255)"`
	QueueID             *string `gorm:"type:varchar(255)"`
	ExpireTime          *string `gorm:"type:varchar(255)"`
	TrafficLimit        *string `gorm:"type:varchar(255)"`
	DownloadBandwidth   *string `gorm:"type:varchar(255)"`
	UploadBandwidth     *string `gorm:"type:varchar(255)"`
	DownloadUsage       string  `gorm:"type:varchar(255);not null;default:'0'"`
	UploadUsage         string  `gorm:"type:varchar(255);not null;default:'0'"`
}
