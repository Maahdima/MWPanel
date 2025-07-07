package model

import "gorm.io/gorm"

type Interface struct {
	gorm.Model
	InterfaceID string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Disabled    bool    `gorm:"type:boolean;not null;default:false"`
	Comment     *string `gorm:"type:varchar(255)"`
	Name        string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	ListenPort  string  `gorm:"type:varchar(10);not null"`
}
