package model

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt uint64
	UpdatedAt uint64
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *Model) BeforeUpdate(db *gorm.DB) error {
	db.UpdateColumn("UpdatedAt", time.Now().Unix())
	return nil
}

func (m *Model) BeforeCreate(db *gorm.DB) error {
	if m.UpdatedAt == 0 {
		db.UpdateColumn("UpdatedAt", time.Now().Unix())
	}

	db.UpdateColumn("CreatedAt", time.Now().Unix())
	return nil
}
