package model

type IPPool struct {
	Model
	Name    string `gorm:"type:varchar(255);not null;uniqueIndex"`
	StartIP string `gorm:"type:varchar(15);not null;uniqueIndex"`
	EndIP   string `gorm:"type:varchar(15);not null;uniqueIndex"`

	InterfaceID uint `gorm:"not null;uniqueIndex"`
}
