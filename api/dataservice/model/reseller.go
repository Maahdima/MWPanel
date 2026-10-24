package model

// Reseller is a tenant, not a separate installation. Login credentials remain in
// Admin so JWT and TOTP keep the existing authentication path.
type Reseller struct {
	Model
	Name     string `gorm:"type:varchar(255);not null"`
	IsActive bool   `gorm:"not null;default:true"`
}
