package schema

type ServerStatus string

var (
	AvailableServer    ServerStatus = "available"
	NotAvailableServer ServerStatus = "not_available"
)

type CreateServerRequest struct {
	Comment   *string `json:"comment,omitempty"`
	Name      string  `json:"name" validate:"required"`
	IPAddress string  `json:"ip_address" validate:"required"`
	APIPort   string  `json:"api_port" validate:"required"`
	Username  string  `json:"username" validate:"required"`
	Password  string  `json:"password" validate:"required"`
}

type UpdateServerRequest struct {
	Comment   *string `json:"comment,omitempty"`
	Name      *string `json:"name,omitempty"`
	IPAddress *string `json:"ip_address,omitempty"`
	APIPort   *string `json:"api_port,omitempty"`
	Username  *string `json:"username,omitempty"`
	Password  *string `json:"password,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

type ServerResponse struct {
	Id        uint         `json:"id"`
	Comment   *string      `json:"comment"`
	Name      string       `json:"name"`
	IPAddress string       `json:"ip_address"`
	APIPort   string       `json:"api_port"`
	IsActive  bool         `json:"is_active"`
	Status    ServerStatus `json:"status"`
}

type ServerStatsResponse struct {
	TotalServers  int `json:"total_servers"`
	ActiveServers int `json:"active_servers"`
}
