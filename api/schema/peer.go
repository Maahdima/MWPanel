package schema

type PeerStatus string

var (
	ActivePeer    PeerStatus = "active"
	InactivePeer  PeerStatus = "inactive"
	ExpiredPeer   PeerStatus = "expired"
	SuspendedPeer PeerStatus = "suspended"
)

type CreatePeerRequest struct {
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

type UpdatePeerRequest struct {
	Disabled            *bool   `json:"disabled"`
	Comment             *string `json:"comment"`
	Name                *string `json:"name"`
	AllowedAddress      *string `json:"allowed_address"`
	PresharedKey        *string `json:"preshared_key"`
	PersistentKeepAlive *string `json:"persistent_keepalive"`
	ExpireTime          *string `json:"expire_time"`
	TrafficLimit        *string `json:"traffic_limit"`
	DownloadBandwidth   *string `json:"download_bandwidth"`
	UploadBandwidth     *string `json:"upload_bandwidth"`
}

type PeerKeyResponse struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type PeerResponse struct {
	Id                uint         `json:"id"`
	Disabled          bool         `json:"disabled"`
	Comment           *string      `json:"comment"`
	Name              string       `json:"name"`
	Interface         string       `json:"interface"`
	AllowedAddress    string       `json:"allowed_address"`
	TrafficLimit      *string      `json:"traffic_limit"`
	ExpireTime        *string      `json:"expire_time"`
	DownloadBandwidth *string      `json:"download_bandwidth"`
	UploadBandwidth   *string      `json:"upload_bandwidth"`
	Status            []PeerStatus `json:"status"`
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
