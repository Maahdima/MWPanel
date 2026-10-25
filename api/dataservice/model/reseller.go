package model

import "gorm.io/gorm"

// Reseller is a tenant, not a separate installation. Credentials remain in Admin.
type Reseller struct {
	AdminSuspended     bool   `gorm:"not null;default:false"`
	AdminSuspendReason string `gorm:"type:text"`
	Model
	Name                 string `gorm:"type:varchar(255);not null"`
	IsActive             bool   `gorm:"not null;default:true"`
	DefaultInterfaceID   *uint
	MaxActivePeers       *int64
	MaxPeerTrafficBytes  *int64
	MaxTotalTrafficBytes *int64
	MaxDurationDays      *int64
}

type ResellerResource struct {
	Model
	ResellerID  uint      `gorm:"not null;uniqueIndex:uidx_reseller_interface"`
	Reseller    Reseller  `gorm:"constraint:OnDelete:RESTRICT"`
	InterfaceID uint      `gorm:"not null;uniqueIndex:uidx_reseller_interface"`
	Interface   Interface `gorm:"constraint:OnDelete:RESTRICT"`
}

type PeerOwnershipEvent struct {
	ConfiguredDurationDays *int64
	ID                     string `gorm:"type:varchar(36);primaryKey"`
	CreatedAt              int64  `gorm:"autoCreateTime"`
	PeerID                 uint   `gorm:"not null;index"`
	PreviousResellerID     *uint
	ResellerID             *uint
	ActorID                uint   `gorm:"not null"`
	Reason                 string `gorm:"type:text;not null"`
}

func (*PeerOwnershipEvent) BeforeUpdate(*gorm.DB) error { return ErrImmutableAccounting }
func (*PeerOwnershipEvent) BeforeDelete(*gorm.DB) error { return ErrImmutableAccounting }
