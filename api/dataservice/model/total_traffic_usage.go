package model

const TotalTrafficUsageSingletonID uint = 1

type TotalTrafficUsage struct {
	Model
	TotalUsage int64 `gorm:"type:bigint;not null;default:0"` // in bytes
}
