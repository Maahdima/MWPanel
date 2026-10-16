package schema

type OrderResponse struct {
	ID            uint    `json:"id"`
	TrackingCode  string  `json:"tracking_code"`
	ChatID        string  `json:"chat_id"`
	Username      *string `json:"username"`
	Kind          string  `json:"kind"`
	ConfigName    string  `json:"config_name"`
	PlanTitle     string  `json:"plan_title"`
	PlanName      string  `json:"plan_name"`
	TrafficGB     int     `json:"traffic_gb"`
	DurationDays  int     `json:"duration_days"`
	AmountTomans  int64   `json:"amount_tomans"`
	PayerCard     string  `json:"payer_card"`
	Status        string  `json:"status"`
	ExpiresAt     int64   `json:"expires_at"`
	PaidAt        *int64  `json:"paid_at"`
	CreatedAt     uint64  `json:"created_at"`
	FailureReason *string `json:"failure_reason"`
	Manual        bool    `json:"manual"`
	ConfigID      *uint   `json:"config_id"`
	PeerName      *string `json:"peer_name"`
	PeerExpire    *string `json:"peer_expire"`
}
