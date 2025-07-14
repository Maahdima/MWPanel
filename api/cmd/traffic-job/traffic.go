package traffic

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"mikrotik-wg-go/adaptor/mikrotik"
	"mikrotik-wg-go/dataservice/model"
	"mikrotik-wg-go/utils"
	"strconv"
	"sync"
)

type Calculator struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	mu              *sync.Mutex // avoid race condition between traffic job and reset
	logger          *zap.Logger
}

func NewTrafficCalculator(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor) *Calculator {
	return &Calculator{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		mu:              &sync.Mutex{},
		logger:          zap.L().Named("TrafficCalculatorJob"),
	}
}

func (c *Calculator) CalculateTraffic() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var peers []model.Peer
	if err := c.db.Find(&peers).Error; err != nil {
		c.logger.Error("Failed to fetch peers from database", zap.Error(err))
		return
	}

	const maxCounter = 4294967296 // mikrotik 32-bit counter bug in wg peers (2^32)

	for _, peer := range peers {
		wgPeer, err := c.mikrotikAdaptor.FetchWgPeer(context.Background(), peer.PeerID)
		if err != nil {
			c.logger.Error("Failed to fetch wireguard peer", zap.String("peerID", peer.PeerID), zap.Error(err))
			continue
		}

		currentTx := utils.ParseStringToInt(*wgPeer.TransferTx)
		currentRx := utils.ParseStringToInt(*wgPeer.TransferRx)

		deltaTx := calculateDelta(peer.LastTx, currentTx, maxCounter)
		deltaRx := calculateDelta(peer.LastRx, currentRx, maxCounter)

		peer.DownloadUsage += deltaTx
		peer.UploadUsage += deltaRx

		peer.LastTx = currentTx
		peer.LastRx = currentRx

		totalBytes := peer.DownloadUsage + peer.UploadUsage
		if peer.TrafficLimit > 0 && totalBytes > peer.TrafficLimit {
			c.logger.Warn("Peer traffic limit exceeded", zap.String("peerID", peer.PeerID))

			peer.Disabled = true
			_, err := c.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{
				Disabled: utils.Ptr(strconv.FormatBool(true)),
			})
			if err != nil {
				c.logger.Error("Failed to disable peer on Mikrotik", zap.String("peerID", peer.PeerID), zap.Error(err))
			}
		}

		if err := c.db.Save(&peer).Error; err != nil {
			c.logger.Error("Failed to update peer usage in database", zap.String("peerID", peer.PeerID), zap.Error(err))
		}
	}

	c.logger.Info("Traffic calculation job completed")
}

func (c *Calculator) ResetPeerUsage(id uint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx := context.Background()

	var peer model.Peer
	if err := c.db.Where("id = ?", id).First(&peer).Error; err != nil {
		c.logger.Error("Failed to find peer in DB", zap.Uint("id", id), zap.Error(err))
		return err
	}

	wgPeer, err := c.mikrotikAdaptor.FetchWgPeer(ctx, peer.PeerID)
	if err != nil {
		c.logger.Error("Failed to fetch peer from Mikrotik", zap.String("peerID", peer.PeerID), zap.Error(err))
		return err
	}

	currentTx := utils.ParseStringToInt(*wgPeer.TransferTx)
	currentRx := utils.ParseStringToInt(*wgPeer.TransferRx)

	err = c.db.Transaction(func(tx *gorm.DB) error {
		peer.DownloadUsage = 0
		peer.UploadUsage = 0
		peer.LastTx = currentTx
		peer.LastRx = currentRx

		if err := tx.Save(&peer).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.logger.Error("Failed to reset peer usage", zap.String("peerID", peer.PeerID), zap.Error(err))
		return err
	}

	c.logger.Info("Peer usage reset successfully", zap.String("peerID", peer.PeerID))

	return nil
}

func calculateDelta(prev, current, maxCounter int64) int64 {
	if current >= prev {
		return current - prev
	}
	return (maxCounter - prev) + current
}
