package schema

type InterfaceResponse struct {
	InterfaceID string `json:"interface_id"`
	Name        string `json:"name"`
	ListenPort  string `json:"listen_port"`
}

type InterfacesDataResponse struct {
	TotalServers  int `json:"total_servers"`
	ActiveServers int `json:"active_servers"`
}
