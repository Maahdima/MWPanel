package schema

type PeerStatus string

var (
	ActivePeer    PeerStatus = "active"
	InactivePeer  PeerStatus = "inactive"
	ExpiredPeer   PeerStatus = "expired"
	SuspendedPeer PeerStatus = "suspended"
)

type CreatePeerRequest struct {
	Comment             *string `json:"comment,omitempty"`
	Name                string  `json:"name" validate:"required"`
	InterfaceId         string  `json:"interface_id" validate:"required"`
	Interface           string  `json:"interface_name" validate:"required"`
	PrivateKey          string  `json:"private_key" validate:"required"`
	PublicKey           string  `json:"public_key" validate:"required"`
	AllowedAddress      string  `json:"allowed_address" validate:"required"`
	PresharedKey        *string `json:"preshared_key,omitempty"`
	PersistentKeepAlive *string `json:"persistent_keepalive"`
	Endpoint            string  `json:"endpoint" validate:"required"`
	ExpireTime          *string `json:"expire_time,omitempty"`
	TrafficLimit        *string `json:"traffic_limit,omitempty"`
	DownloadBandwidth   *string `json:"download_bandwidth,omitempty"`
	UploadBandwidth     *string `json:"upload_bandwidth,omitempty"`
}

type UpdatePeerRequest struct {
	Disabled            *bool   `json:"disabled,omitempty"`
	Comment             *string `json:"comment,omitempty"`
	Name                *string `json:"name,omitempty"`
	AllowedAddress      *string `json:"allowed_address,omitempty"`
	PresharedKey        *string `json:"preshared_key,omitempty"`
	PersistentKeepAlive *string `json:"persistent_keepalive,omitempty"`
	ExpireTime          *string `json:"expire_time,omitempty"`
	TrafficLimit        *string `json:"traffic_limit,omitempty"`
	DownloadBandwidth   *string `json:"download_bandwidth,omitempty"`
	UploadBandwidth     *string `json:"upload_bandwidth,omitempty"`
}

type PeerKeyResponse struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type PeerDetailsResponse struct {
	Name          string  `json:"name"`
	TrafficLimit  *string `json:"traffic_limit"`
	ExpireTime    *string `json:"expire_time"`
	DownloadUsage string  `json:"download_usage"`
	UploadUsage   string  `json:"upload_usage"`
	TotalUsage    string  `json:"total_usage"`
	UsagePercent  *string `json:"usage_percent"`
}

type PeerResponse struct {
	Id                uint         `json:"id"`
	UUID              string       `json:"uuid"`
	Disabled          bool         `json:"disabled"`
	Comment           *string      `json:"comment"`
	Name              string       `json:"name"`
	Interface         string       `json:"interface"`
	AllowedAddress    string       `json:"allowed_address"`
	TrafficLimit      *string      `json:"traffic_limit"`
	ExpireTime        *string      `json:"expire_time"`
	DownloadBandwidth *string      `json:"download_bandwidth"`
	UploadBandwidth   *string      `json:"upload_bandwidth"`
	TotalUsage        string       `json:"total_usage"`
	Status            []PeerStatus `json:"status"`
	IsShared          bool         `json:"is_shared"`
}

type PeerStatsResponse struct {
	RecentOnlinePeers *[]RecentOnlinePeers `json:"recent_online_peers"`
	TotalPeers        int                  `json:"total_peers"`
	OnlinePeers       int                  `json:"online_peers"`
}

type RecentOnlinePeers struct {
	Name     string `json:"name"`
	LastSeen string `json:"last_seen"`
}
