package mikrotik

import (
	"context"

	"github.com/maahdima/mwp/api/common"
)

type Interface struct {
	ID        string `json:".id,omitempty"`
	Name      string `json:"name,omitempty"`
	ActualMTU string `json:"actual-mtu,omitempty"`
	Disabled  string `json:"disabled,omitempty"`
	MTU       string `json:"mtu,omitempty"`
	TxByte    string `json:"tx-byte,omitempty"`
	RxByte    string `json:"rx-byte,omitempty"`
	Running   string `json:"running,omitempty"`
	Type      string `json:"type,omitempty"`
}

func (a *Adaptor) FetchInterface(c context.Context, interfaceID string) (*Interface, error) {
	var iface Interface

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Get(
		c,
		common.InterfacePath+"/"+interfaceID,
		&iface,
	)
	if err != nil {
		return nil, err
	}

	return &iface, nil
}
