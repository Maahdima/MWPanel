package model

type Traffic struct {
	Model
	InterfaceID   uint  `gorm:"index;not null"`
	DownloadUsage int64 `gorm:"type:bigint;not null;default:0"` // in bytes
	UploadUsage   int64 `gorm:"type:bigint;not null;default:0"` // in bytes
	TotalUsage    int64 `gorm:"type:bigint;not null;default:0"` // in bytes
}
