package schema

type SyncInterfacePreviewResponse struct {
	ID         string  `json:"id"`
	Disabled   bool    `json:"disabled"`
	Comment    *string `json:"comment"`
	Name       string  `json:"name"`
	ListenPort string  `json:"listen_port"`
	MTU        string  `json:"mtu"`
	IsRunning  bool    `json:"is_running"`
}

type SyncPeerPreviewResponse struct {
	ID             string  `json:"id"`
	Disabled       bool    `json:"disabled"`
	Comment        *string `json:"comment"`
	Name           string  `json:"name"`
	Interface      string  `json:"interface"`
	AllowedAddress string  `json:"allowed_address"`
}

type SyncInterfacesRequest struct {
	InterfaceIDs []string `json:"interface_ids"`
}

type SyncPeersRequest struct {
	PeerIDs []string `json:"peer_ids"`
}
