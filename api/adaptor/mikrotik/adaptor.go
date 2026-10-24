package mikrotik

import (
	"errors"
	"github.com/maahdima/mwp/api/utils/httphelper"
	"go.uber.org/zap"

	"github.com/maahdima/mwp/api/common"
)

type Adaptor struct {
	boundClient *httphelper.Client
	mwpClient   *common.MwpClient
	logger      *zap.Logger
}

// NewAdaptor creates a new instance of the Mikrotik adaptor
func NewAdaptor(mwpClients *common.MwpClient) *Adaptor {
	return &Adaptor{
		mwpClient: mwpClients,
		logger:    zap.L().Named("MikrotikAdaptor"),
	}
}

// ForServer returns an independent adaptor bound to one authorized server. It
// does not mutate the shared adaptor or client registry, so calls can overlap.
func (a *Adaptor) ForServer(serverID uint) (*Adaptor, error) {
	client, err := a.mwpClient.ClientForServer(serverID)
	if err != nil {
		return nil, err
	}
	return &Adaptor{mwpClient: a.mwpClient, boundClient: client, logger: a.logger}, nil
}

// ForConnection is used for preflight before a server is persisted.
func (a *Adaptor) ForConnection(client *httphelper.Client) *Adaptor {
	return &Adaptor{mwpClient: a.mwpClient, boundClient: client, logger: a.logger}
}

func (a *Adaptor) client() (*httphelper.Client, error) {
	if a.boundClient != nil {
		return a.boundClient, nil
	}
	if a.mwpClient != nil {
		if client := a.mwpClient.GetClient(nil); client != nil {
			return client, nil
		}
	}
	return nil, errors.New("no active MikroTik server is configured")
}
