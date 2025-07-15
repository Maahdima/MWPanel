package model

import "gorm.io/gorm"

type Server struct {
	gorm.Model
	Comment   *string `gorm:"type:varchar(255)"`
	Name      string  `gorm:"type:varchar(64);uniqueIndex;not null;"`
	IPAddress string  `gorm:"type:varchar(64);uniqueIndex;not null;"`
	APIPort   int     `gorm:"not null;default:80;"`
	Username  string  `gorm:"type:varchar(64);not null;"`
	Password  string  `gorm:"type:varchar(64);not null;"`
	IsActive  bool    `gorm:"not null;default:true;"`
}
