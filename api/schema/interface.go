package schema

type InterfaceResponse struct {
	Id          uint    `json:"id"`
	InterfaceID string  `json:"interface_id"`
	Disabled    bool    `json:"disabled"`
	Comment     *string `json:"comment"`
	Name        string  `json:"name"`
	ListenPort  string  `json:"listen_port"`
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

type InterfacesDataResponse struct {
	TotalServers  int `json:"total_servers"`
	ActiveServers int `json:"active_servers"`
}
