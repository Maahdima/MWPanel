package model

type Interface struct {
	Model
	InterfaceID string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Disabled    bool    `gorm:"type:boolean;not null;default:false"`
	Comment     *string `gorm:"type:varchar(255)"`
	Name        string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	PrivateKey  string  `gorm:"type:varchar(255);not null"`
	PublicKey   string  `gorm:"type:varchar(255);not null"`
	ListenPort  string  `gorm:"type:varchar(10);not null"`

	IPPool *IPPool `gorm:"foreignKey:InterfaceID"`
}
