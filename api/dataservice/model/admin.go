package model

type Admin struct {
	Model
	Username string `gorm:"type:varchar(64);uniqueIndex;not null;"`
	Password string `gorm:"type:varchar(128);not null;"`
	IsActive bool   `gorm:"not null;default:true;"`
}
