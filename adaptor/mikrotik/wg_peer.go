package mikrotik

import (
	"context"
	"go.uber.org/zap"
)

type WireGuardPeer struct {
	ID                     *string `json:".id,omitempty"`
	Disabled               *string `json:"disabled,omitempty"`
	Comment                *string `json:"comment,omitempty"`
	AllowedAddress         *string `json:"allowed-address,omitempty"`
	PersistentKeepAlive    *string `json:"persistent-keepalive,omitempty"`
	Interface              *string `json:"interface,omitempty"`
	Name                   *string `json:"name,omitempty"`
	PresharedKey           *string `json:"preshared-key,omitempty"`
	PublicKey              *string `json:"public-key,omitempty"`
	ClientEndpoint         *string `json:"client-endpoint,omitempty"`
	CurrentEndpointAddress *string `json:"current-endpoint-address,omitempty"`
	CurrentEndpointPort    *string `json:"current-endpoint-port,omitempty"`
	LastHandshake          *string `json:"last-handshake,omitempty"`
}

func (a *Adaptor) FetchWgPeers(c context.Context) (*[]WireGuardPeer, error) {
	var wgPeers []WireGuardPeer

	err := a.httpClient.Get(
		c,
		WGPeerPath,
		&wgPeers,
	)
	if err != nil {
		a.logger.Error("failed to get wireguard peers", zap.Error(err))
		return nil, err
	}

	return &wgPeers, nil
}

func (a *Adaptor) FetchWgPeer(c context.Context, peerID string) (*WireGuardPeer, error) {
	var wgPeer WireGuardPeer

	err := a.httpClient.Get(
		c,
		WGPeerPath+"/"+peerID,
		&wgPeer,
	)
	if err != nil {
		a.logger.Error("failed to get wireguard peer", zap.Error(err))
		return nil, err
	}

	return &wgPeer, nil
}

func (a *Adaptor) CreateWgPeer(c context.Context, peer WireGuardPeer) (*WireGuardPeer, error) {
	var createdPeer WireGuardPeer

	err := a.httpClient.Put(
		c,
		WGPeerPath,
		peer,
		&createdPeer,
	)
	if err != nil {
		return nil, err
	}

	return &createdPeer, nil
}

func (a *Adaptor) UpdateWgPeer(c context.Context, peerID string, peer WireGuardPeer) (*WireGuardPeer, error) {
	var updatedPeer WireGuardPeer

	err := a.httpClient.Patch(
		c,
		WGPeerPath+"/"+peerID,
		peer,
		&updatedPeer,
	)
	if err != nil {
		return nil, err
	}

	return &updatedPeer, nil
}

func (a *Adaptor) DeleteWgPeer(c context.Context, peerID string) error {
	err := a.httpClient.Delete(
		c,
		WGPeerPath+"/"+peerID,
		nil,
	)
	if err != nil {
		a.logger.Error("failed to delete wireguard peer", zap.Error(err))
		return err
	}

	return nil
}
