package model

type IPPool struct {
	Model
	ServerID uint   `gorm:"not null;default:0;index;uniqueIndex:uidx_pool_router_name;uniqueIndex:uidx_pool_router_start;uniqueIndex:uidx_pool_router_end"`
	Name     string `gorm:"type:varchar(255);not null;uniqueIndex:uidx_pool_router_name"`
	StartIP  string `gorm:"type:varchar(15);not null;uniqueIndex:uidx_pool_router_start"`
	EndIP    string `gorm:"type:varchar(15);not null;uniqueIndex:uidx_pool_router_end"`

	InterfaceID uint `gorm:"not null;uniqueIndex"`
}
