package model

// Deposit is one bank message received from bale-notifier.
// Fingerprint stops the same webhook delivery from charging a config twice.
type Deposit struct {
	Model
	Fingerprint  string `gorm:"type:varchar(64);uniqueIndex;not null"`
	RawText      string `gorm:"type:text;not null"`
	AmountTomans *int64
	OrderID      *uint  `gorm:"index"`
	Order        *Order `gorm:"foreignKey:OrderID;references:ID"`
	Matched      bool   `gorm:"not null;default:false"`
	ReceivedAt   int64  `gorm:"not null"`
}

func (Deposit) TableName() string { return "vpn_deposits" }
