package mikrotik

import (
	"context"
	"go.uber.org/zap"
)

type WireGuardInterface struct {
	ID         *string `json:".id,omitempty"`
	Disabled   *string `json:"disabled,omitempty"`
	Comment    *string `json:"comment,omitempty"`
	ListenPort *string `json:"listen-port,omitempty"`
	MTU        *string `json:"mtu,omitempty"`
	Name       *string `json:"name,omitempty"`
	PrivateKey *string `json:"private-key,omitempty"`
	PublicKey  *string `json:"public-key,omitempty"`
	Running    *string `json:"running,omitempty"`
}

func (a *Adaptor) FetchWgInterfaces(c context.Context) (*[]WireGuardInterface, error) {
	var wgInterfaces []WireGuardInterface

	err := a.httpClient.Get(
		c,
		WGInterfacePath,
		&wgInterfaces,
	)
	if err != nil {
		return nil, err
	}

	return &wgInterfaces, nil
}

func (a *Adaptor) FetchWgInterface(c context.Context, interfaceID string) (*WireGuardInterface, error) {
	var wgInterface WireGuardInterface

	err := a.httpClient.Get(
		c,
		WGInterfacePath+"/"+interfaceID,
		&wgInterface,
	)
	if err != nil {
		return nil, err
	}

	return &wgInterface, nil
}

func (a *Adaptor) CreateWgInterface(c context.Context, wgInterface WireGuardInterface) (*WireGuardInterface, error) {
	var createdInterface WireGuardInterface

	err := a.httpClient.Put(
		c,
		WGInterfacePath,
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

	err := a.httpClient.Patch(
		c,
		WGInterfacePath+"/"+interfaceID,
		wgInterface,
		&updatedInterface,
	)
	if err != nil {
		return nil, err
	}

	return &updatedInterface, nil
}

func (a *Adaptor) DeleteWgInterface(c context.Context, interfaceID string) error {
	err := a.httpClient.Delete(
		c,
		WGInterfacePath+"/"+interfaceID,
		nil,
	)
	if err != nil {
		a.logger.Error("failed to delete wireguard peer", zap.Error(err))
		return err
	}

	return nil
}
