package model

type Plan struct {
	Model
	Name         string  `gorm:"type:varchar(255);not null;default:''"`
	FaName       string  `gorm:"column:fa_name;type:varchar(255);not null;default:''"`
	Title        string  `gorm:"type:varchar(255);not null"`
	TrafficGB    int     `gorm:"not null;uniqueIndex:idx_vpn_plan_offer"`
	DurationDays int     `gorm:"not null;uniqueIndex:idx_vpn_plan_offer"`
	Enabled      bool    `gorm:"not null;default:true"`
	SortOrder    int     `gorm:"not null;default:0"`
	Orders       []Order `gorm:"foreignKey:PlanID;references:ID"`
}

func (Plan) TableName() string { return "vpn_plans" }
