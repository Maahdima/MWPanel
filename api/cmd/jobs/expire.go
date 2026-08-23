package traffic

import (
	"context"
	"strconv"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"

	"go.uber.org/zap"
)

func (c *Calculator) ExpireOverduePeers() {
	c.mu.Lock()
	defer c.mu.Unlock()

	peers, err := c.fetchPeers()
	if err != nil {
		return
	}

	now := time.Now()
	expiredCount := 0
	for _, peer := range peersDueForExpiry(peers, now) {
		if err := c.disableExpiredPeer(peer); err != nil {
			continue
		}
		expiredCount++
	}

	if expiredCount > 0 {
		c.logger.Info("Disabled expired peers", zap.Int("count", expiredCount))
	}
}

func peersDueForExpiry(peers []model.Peer, now time.Time) []model.Peer {
	due := make([]model.Peer, 0)
	for _, peer := range peers {
		if peer.Disabled || !utils.IsPeerExpired(peer.ExpireTime, now) {
			continue
		}
		due = append(due, peer)
	}
	return due
}

func (c *Calculator) disableExpiredPeer(peer model.Peer) error {
	_, err := c.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{
		Disabled: strconv.FormatBool(true),
	})
	if err != nil {
		c.logger.Error("Failed to disable expired peer on Mikrotik", zap.String("peerID", peer.PeerID), zap.Error(err))
		return err
	}

	if peer.QueueID != nil {
		_, err := c.mikrotikAdaptor.UpdateSimpleQueue(context.Background(), *peer.QueueID, mikrotik.Queue{
			Disabled: strconv.FormatBool(true),
		})
		if err != nil {
			c.logger.Error("Failed to disable queue for expired peer", zap.String("peerID", peer.PeerID), zap.Error(err))
		}
	}

	if err := c.db.Model(&model.Peer{}).Where("id = ?", peer.ID).Update("disabled", true).Error; err != nil {
		c.logger.Error("Failed to mark expired peer as disabled", zap.String("peerID", peer.PeerID), zap.Error(err))
		return err
	}

	c.logger.Info("Disabled expired peer", zap.String("peerID", peer.PeerID), zap.String("name", peer.Name))
	return nil
}
