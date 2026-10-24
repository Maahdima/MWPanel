package model

type Interface struct {
	Model
	ServerID          uint    `gorm:"not null;default:0;index;uniqueIndex:uidx_interface_router_id;uniqueIndex:uidx_interface_router_name"`
	InterfaceID       string  `gorm:"type:varchar(255);uniqueIndex:uidx_interface_router_id;not null"`
	Disabled          bool    `gorm:"type:boolean;not null;default:false"`
	Comment           *string `gorm:"type:varchar(255)"`
	Name              string  `gorm:"type:varchar(255);uniqueIndex:uidx_interface_router_name;not null"`
	PrivateKey        string  `gorm:"type:varchar(255);not null"`
	PublicKey         string  `gorm:"type:varchar(255);not null"`
	ListenPort        string  `gorm:"type:varchar(10);not null"`
	LastTx            int64   `gorm:"type:bigint;not null;default:0"`
	LastRx            int64   `gorm:"type:bigint;not null;default:0"`
	DailyTrafficReady bool    `gorm:"type:boolean;not null;default:false"`

	IPPool *IPPool `gorm:"foreignKey:InterfaceID"`
}
