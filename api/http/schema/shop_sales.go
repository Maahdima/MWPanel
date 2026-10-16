package schema

type ShopStatus struct {
	Enabled bool `json:"enabled"`
}

type ShopSales struct {
	AllowNew   bool `json:"allow_new"`
	AllowRenew bool `json:"allow_renew"`
}
