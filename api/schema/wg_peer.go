package schema

type WgPeerRequest struct {
	Comment             *string `json:"comment"`
	Name                string  `json:"name" validate:"required"`
	InterfaceId         string  `json:"interface_id" validate:"required"`
	Interface           string  `json:"interface_name" validate:"required"`
	PrivateKey          string  `json:"private_key" validate:"required"`
	PublicKey           string  `json:"public_key" validate:"required"`
	AllowedAddress      *string `json:"allowed_address"`
	PresharedKey        *string `json:"preshared_key"`
	PersistentKeepAlive *string `json:"persistent_keepalive"`
	Endpoint            string  `json:"endpoint" validate:"required"`
	ExpireTime          *string `json:"expire_time"`
	TrafficLimit        *string `json:"traffic_limit"`
	DownloadBandwidth   *string `json:"download_bandwidth"`
	UploadBandwidth     *string `json:"upload_bandwidth"`
}

type WgPeerResponse struct {
	Id                string  `json:"id"`
	Disabled          string  `json:"disabled"`
	Comment           *string `json:"comment"`
	Name              string  `json:"name"`
	Interface         string  `json:"interface"`
	AllowedAddress    string  `json:"allowed_address"`
	TrafficLimit      *string `json:"traffic_limit"`
	ExpireTime        *string `json:"expire_time"`
	DownloadBandwidth *string `json:"download_bandwidth"`
	UploadBandwidth   *string `json:"upload_bandwidth"`
}

type PeersDataResponse struct {
	RecentOnlinePeers *[]RecentOnlinePeers `json:"recent_online_peers"`
	TotalPeers        int                  `json:"total_peers"`
	OnlinePeers       int                  `json:"online_peers"`
}

type RecentOnlinePeers struct {
	Name     string `json:"name"`
	LastSeen string `json:"last_seen"`
}
