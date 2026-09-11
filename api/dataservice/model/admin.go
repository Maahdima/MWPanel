package model

type Admin struct {
	Model
	Username          string  `gorm:"type:varchar(64);uniqueIndex;not null;"`
	Password          string  `gorm:"type:varchar(128);not null;"`
	IsActive          bool    `gorm:"not null;default:true;"`
	TotpSecret        *string `gorm:"type:varchar(64)"`
	TotpPendingSecret *string `gorm:"type:varchar(64)"`
	TotpEnabled       bool    `gorm:"not null;default:false"`
	TotpRecoveryCodes *string `gorm:"type:text"` // JSON array of bcrypt-hashed recovery codes
}
