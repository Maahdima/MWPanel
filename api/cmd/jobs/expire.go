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
	c.notifyUpcomingExpiries(peers, now)

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

func (c *Calculator) notifyUpcomingExpiries(peers []model.Peer, now time.Time) {
	if c.notifier == nil || !c.notifier.Enabled() {
		return
	}

	for i := range peers {
		peer := &peers[i]
		if peer.Disabled {
			continue
		}

		daysLeft, ok := utils.DaysUntilPeerExpire(peer.ExpireTime, now)
		if !ok || daysLeft < 1 || daysLeft > 3 {
			continue
		}

		updates := map[string]interface{}{}
		c.notifyExpiryThreshold(peer, updates, daysLeft, 3, "expire_first_notify", &peer.ExpireFirstNotify)
		c.notifyExpiryThreshold(peer, updates, daysLeft, 2, "expire_second_notify", &peer.ExpireSecondNotify)
		c.notifyExpiryThreshold(peer, updates, daysLeft, 1, "expire_third_notify", &peer.ExpireThirdNotify)

		if len(updates) == 0 {
			continue
		}

		if err := c.db.Model(&model.Peer{}).Where("id = ?", peer.ID).Updates(updates).Error; err != nil {
			c.logger.Error("Failed to update peer expiry notify flags", zap.String("peerID", peer.PeerID), zap.Error(err))
		}
	}
}

func (c *Calculator) notifyExpiryThreshold(peer *model.Peer, updates map[string]interface{}, daysLeft, threshold int, updateKey string, notified *bool) {
	if daysLeft != threshold || *notified {
		return
	}

	err := c.notifier.NotifyPeerExpiry(context.Background(), peer, daysLeft)
	if err != nil {
		c.logger.Warn("Failed to send peer expiry notification", zap.String("peerID", peer.PeerID), zap.Int("daysLeft", daysLeft), zap.Error(err))
		return
	}

	*notified = true
	updates[updateKey] = true
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
