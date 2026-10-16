package model

const (
	OrderKindNew   = "new"
	OrderKindRenew = "renew"

	OrderPending    = "pending"
	OrderFulfilling = "fulfilling"
	OrderPaid       = "paid"
	OrderFailed     = "failed"
	OrderCancelled  = "cancelled"
	OrderExpired    = "expired"
)

// Order is one Telegram purchase waiting for a bank deposit, or already settled.
type Order struct {
	Model
	ChatID        string  `gorm:"type:varchar(64);not null;index"`
	Username      *string `gorm:"type:varchar(255)"`
	Kind          string  `gorm:"type:varchar(16);not null"`
	ConfigName    string  `gorm:"type:varchar(255);not null"`
	PlanID        uint    `gorm:"not null;index"`
	Plan          *Plan   `gorm:"foreignKey:PlanID;references:ID"`
	ConfigID      *uint   `gorm:"index"`
	Config        *Peer   `gorm:"foreignKey:ConfigID;references:ID"`
	PlanTitle     string  `gorm:"type:varchar(255);not null"`
	TrafficGB     int     `gorm:"not null"`
	DurationDays  int     `gorm:"not null"`
	BaseTomans    int64   `gorm:"not null"`
	AmountTomans  int64   `gorm:"not null;index"`
	PayerCard     string  `gorm:"type:varchar(16);not null;index"`
	Status        string  `gorm:"type:varchar(16);not null;index"`
	ExpiresAt     int64   `gorm:"not null;index"`
	PaidAt        *int64
	NotifiedAt    *int64
	DepositID     *uint
	Deposit       *Deposit `gorm:"foreignKey:DepositID;references:ID"`
	TrackingCode  string   `gorm:"type:varchar(12);index"`
	Manual        bool     `gorm:"not null;default:false"`
	FailureReason *string  `gorm:"type:text"`
}

func (Order) TableName() string { return "vpn_orders" }
