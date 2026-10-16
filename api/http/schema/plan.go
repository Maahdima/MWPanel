package schema

type PlanResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	FaName       string `json:"fa_name"`
	TrafficGB    int    `json:"traffic_gb"`
	DurationDays int    `json:"duration_days"`
	Enabled      bool   `json:"enabled"`
	SortOrder    int    `json:"sort_order"`
	PriceTomans  int64  `json:"price_tomans"`
}

type UpsertPlanRequest struct {
	Name         string `json:"name"`
	FaName       string `json:"fa_name"`
	TrafficGB    int    `json:"traffic_gb"`
	DurationDays int    `json:"duration_days"`
	Enabled      *bool  `json:"enabled"`
	SortOrder    int    `json:"sort_order"`
}
