package model

type ShopSetting struct {
	Model
	AllowNew   bool
	AllowRenew bool
}

func (ShopSetting) TableName() string {
	return "shop_settings"
}
