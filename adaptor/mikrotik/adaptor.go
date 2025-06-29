package mikrotik

import (
	"go.uber.org/zap"
	"mikrotik-wg-go/utils/httphelper"
)

var (
	DeviceResourcePath = "/system/resource"
	WGPeerPath         = "/interface/wireguard/peers"
	WGInterfacePath    = "/interface/wireguard"
	QueuePath          = "/queue/simple"
	SchedulerPath      = "/system/scheduler"
)

type Adaptor struct {
	httpClient *httphelper.Client
	logger     *zap.Logger
}

// NewAdaptor creates a new instance of the Mikrotik adaptor
func NewAdaptor(httpClient *httphelper.Client) *Adaptor {
	return &Adaptor{
		httpClient: httpClient,
		logger:     zap.L().Named("MikrotikAdaptor"),
	}
}
