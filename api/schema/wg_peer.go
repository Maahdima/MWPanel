package schema

type WgPeerRequest struct {
	Comment             *string `json:"comment"`
	Name                string  `json:"name" validate:"required"`
	InterfaceId         string  `json:"interface-id" validate:"required"`
	Interface           string  `json:"interface-name" validate:"required"`
	PrivateKey          string  `json:"private-key" validate:"required"`
	PublicKey           string  `json:"public-key" validate:"required"`
	AllowedAddress      *string `json:"allowed-address"`
	PresharedKey        *string `json:"preshared-key"`
	PersistentKeepAlive *string `json:"persistent-keepalive"`
	Endpoint            string  `json:"endpoint" validate:"required"`
	ExpireTime          *string `json:"expire-time"`
	TrafficLimit        *string `json:"traffic-limit"`
	DownloadBandwidth   *string `json:"download-bandwidth"`
	UploadBandwidth     *string `json:"upload-bandwidth"`
}

type WgPeerResponse struct {
	Id                string  `json:"id"`
	Disabled          string  `json:"disabled"`
	Comment           *string `json:"comment"`
	Name              string  `json:"name"`
	Interface         string  `json:"interface"`
	AllowedAddress    string  `json:"allowed-address"`
	TrafficLimit      *string `json:"traffic-limit"`
	ExpireTime        *string `json:"expire-time"`
	DownloadBandwidth *string `json:"download-bandwidth"`
	UploadBandwidth   *string `json:"upload-bandwidth"`
}
