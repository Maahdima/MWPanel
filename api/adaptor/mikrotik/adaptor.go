package mikrotik

import (
	"go.uber.org/zap"

	"github.com/maahdima/mwp/api/common"
)

type Adaptor struct {
	mwpClients *common.MwpClients
	logger     *zap.Logger
}

// NewAdaptor creates a new instance of the Mikrotik adaptor
func NewAdaptor(mwpClients *common.MwpClients) *Adaptor {
	return &Adaptor{
		mwpClients: mwpClients,
		logger:     zap.L().Named("MikrotikAdaptor"),
	}
}
