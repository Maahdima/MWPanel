package schema

type ResellerSettings struct {
	IntervalDays         *int64 `json:"interval_days"`
	GraceDays            *int64 `json:"grace_days"`
	RecurringFeeRial     *int64 `json:"recurring_fee_rial"`
	Name                 string `json:"name"`
	Mode                 string `json:"billing_mode"`
	RateRial             *int64 `json:"rate_rial"`
	InterfaceIDs         []uint `json:"interface_ids"`
	DefaultInterfaceID   *uint  `json:"default_interface_id"`
	MaxActivePeers       *int64 `json:"max_active_peers"`
	MaxPeerTrafficBytes  *int64 `json:"max_peer_traffic_bytes"`
	MaxTotalTrafficBytes *int64 `json:"max_total_traffic_bytes"`
	MaxDurationDays      *int64 `json:"max_duration_days"`
}

type CreateReseller struct {
	ResellerSettings
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResellerResourceView struct {
	InterfaceID uint   `json:"interface_id"`
	Name        string `json:"name"`
	ServerID    uint   `json:"server_id"`
	PoolID      *uint  `json:"pool_id"`
}

type ResellerView struct {
	MigrationID          *string                `json:"migration_id"`
	IntervalDays         int64                  `json:"interval_days"`
	GraceDays            int64                  `json:"grace_days"`
	RecurringFeeRial     string                 `json:"recurring_fee_rial"`
	AdminSuspended       bool                   `json:"admin_suspended"`
	BillingSuspended     bool                   `json:"billing_suspended"`
	NextInvoiceAt        int64                  `json:"next_invoice_at"`
	ID                   uint                   `json:"id"`
	Username             string                 `json:"username"`
	Name                 string                 `json:"name"`
	IsActive             bool                   `json:"is_active"`
	Mode                 string                 `json:"billing_mode"`
	RateRial             *string                `json:"rate_rial"`
	BalanceRial          string                 `json:"balance_rial"`
	CreditRial           string                 `json:"credit_rial"`
	DebtRial             string                 `json:"debt_rial"`
	ReservedRial         string                 `json:"reserved_rial"`
	ActivePeers          int64                  `json:"active_peers"`
	TotalPeers           int64                  `json:"total_peers"`
	DefaultInterfaceID   *uint                  `json:"default_interface_id"`
	MaxActivePeers       *string                `json:"max_active_peers"`
	MaxPeerTrafficBytes  *string                `json:"max_peer_traffic_bytes"`
	MaxTotalTrafficBytes *string                `json:"max_total_traffic_bytes"`
	MaxDurationDays      *string                `json:"max_duration_days"`
	Resources            []ResellerResourceView `json:"resources"`
}
