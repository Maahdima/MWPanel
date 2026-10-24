package service

import (
	"context"
	"strings"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/shopclient"
)

func (w *WgPeer) SetShopLinks(client *shopclient.Client) {
	if w == nil {
		return
	}
	w.shopLink = client
}

func (w *WgPeer) FindPeerByID(id uint) (model.Peer, error) {
	var peer model.Peer
	err := w.db.First(&peer, "id = ?", id).Error
	return peer, err
}

func (w *WgPeer) FindPeerByUUID(uuid string) (model.Peer, error) {
	var peer model.Peer
	err := w.db.Where("uuid = ?", strings.TrimSpace(uuid)).First(&peer).Error
	return peer, err
}

func (w *WgPeer) PeerNameTaken(name string) (bool, error) {
	var count int64
	err := w.db.Model(&model.Peer{}).Where("name = ?", strings.TrimSpace(name)).Count(&count).Error
	return count > 0, err
}

func (w *WgPeer) PeerOnline(peer model.Peer) bool {
	if w.mikrotikAdaptor == nil || strings.TrimSpace(peer.PeerID) == "" {
		return false
	}
	bound, err := w.ForServer(peer.ServerID)
	if err != nil {
		return false
	}
	mtPeer, err := bound.mikrotikAdaptor.FetchWgPeer(context.Background(), peer.PeerID)
	if err != nil {
		return false
	}
	_, online, err := w.handshakeData(mtPeer)
	return err == nil && online
}
