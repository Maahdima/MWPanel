package mikrotik

import (
	"context"
)

type WireGuardInterface struct {
	ID         string `json:".id,omitempty"`
	Disabled   string `json:"disabled,omitempty"`
	Comment    string `json:"comment,omitempty"`
	ListenPort string `json:"listen-port"`
	MTU        string `json:"mtu"`
	Name       string `json:"name"`
	PrivateKey string `json:"private-key"`
	PublicKey  string `json:"public-key"`
	Running    string `json:"running"`
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
