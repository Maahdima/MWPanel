package mikrotik

import (
	"go.uber.org/zap"

	"github.com/maahdima/mwp/api/common"
)

type Adaptor struct {
	mwpClient *common.MwpClient
	logger    *zap.Logger
}

// NewAdaptor creates a new instance of the Mikrotik adaptor
func NewAdaptor(mwpClients *common.MwpClient) *Adaptor {
	return &Adaptor{
		mwpClient: mwpClients,
		logger:    zap.L().Named("MikrotikAdaptor"),
	}
}
