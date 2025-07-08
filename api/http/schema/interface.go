package schema

type InterfaceResponse struct {
	Id          uint    `json:"id"`
	InterfaceID string  `json:"interface_id"`
	Disabled    bool    `json:"disabled"`
	Comment     *string `json:"comment"`
	Name        string  `json:"name"`
	ListenPort  string  `json:"listen_port"`
	MTU         string  `json:"mtu"`
	Status      string  `json:"status"`
}

type CreateInterfaceRequest struct {
	Comment    *string `json:"comment,omitempty"`
	Name       string  `json:"name" validate:"required"`
	ListenPort string  `json:"listen_port" validate:"required"`
}

type UpdateInterfaceRequest struct {
	Disabled *bool   `json:"disabled,omitempty"`
	Comment  *string `json:"comment,omitempty"`
	Name     *string `json:"name,omitempty"`
}

type InterfaceStatsResponse struct {
	TotalInterfaces  int `json:"total_interfaces"`
	ActiveInterfaces int `json:"active_interfaces"`
}
