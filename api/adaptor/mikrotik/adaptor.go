package mikrotik

import (
	"github.com/maahdima/mwp/api/utils/httphelper"

	"go.uber.org/zap"
)

var (
	DeviceInfoPath     = "/system/resource"
	DeviceIdentityPath = "/system/identity"
	DeviceDnsPath      = "/ip/dns"
	DeviceIPv4Path     = "/ip/address"
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
