package mikrotik

import (
	"context"

	"go.uber.org/zap"

	"github.com/maahdima/mwp/api/common"
)

type WireGuardInterface struct {
	ID         string  `json:".id"`
	Disabled   string  `json:"disabled,omitempty"`
	Comment    *string `json:"comment,omitempty"`
	ListenPort string  `json:"listen-port,omitempty"`
	MTU        string  `json:"mtu,omitempty"`
	Name       string  `json:"name"`
	PrivateKey string  `json:"private-key,omitempty"`
	PublicKey  string  `json:"public-key,omitempty"`
	Running    *string `json:"running,omitempty"`
}

func (a *Adaptor) FetchWgInterfaces(c context.Context) ([]WireGuardInterface, error) {
	var wgInterfaces []WireGuardInterface

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Get(
		c,
		common.WGInterfacePath,
		&wgInterfaces,
	)
	if err != nil {
		return nil, err
	}

	return wgInterfaces, nil
}

func (a *Adaptor) FetchWgInterface(c context.Context, interfaceID string) (*WireGuardInterface, error) {
	var wgInterface WireGuardInterface

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Get(
		c,
		common.WGInterfacePath+"/"+interfaceID,
		&wgInterface,
	)
	if err != nil {
		return nil, err
	}

	return &wgInterface, nil
}

func (a *Adaptor) CreateWgInterface(c context.Context, wgInterface WireGuardInterface) (*WireGuardInterface, error) {
	var createdInterface WireGuardInterface

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Put(
		c,
		common.WGInterfacePath,
		wgInterface,
		&createdInterface,
	)
	if err != nil {
		return nil, err
	}

	return &createdInterface, nil
}

func (a *Adaptor) UpdateWgInterface(c context.Context, interfaceID string, wgInterface WireGuardInterface) (*WireGuardInterface, error) {
	var updatedInterface WireGuardInterface

	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Patch(
		c,
		common.WGInterfacePath+"/"+interfaceID,
		wgInterface,
		&updatedInterface,
	)
	if err != nil {
		return nil, err
	}

	return &updatedInterface, nil
}

func (a *Adaptor) DeleteWgInterface(c context.Context, interfaceID string) error {
	httpClient := a.mwpClients.GetClient(nil)

	err := httpClient.Delete(
		c,
		common.WGInterfacePath+"/"+interfaceID,
		nil,
	)
	if err != nil {
		a.logger.Error("failed to delete wireguard peer", zap.Error(err))
		return err
	}

	return nil
}
