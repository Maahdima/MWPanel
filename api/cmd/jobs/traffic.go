package traffic

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
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

func (c *Calculator) CalculatePeerTraffic() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var peers []model.Peer
	if err := c.db.Find(&peers).Error; err != nil {
		c.logger.Error("Failed to fetch peers from database", zap.Error(err))
		return
	}

	if len(peers) == 0 {
		c.logger.Info("No peers found, skipping traffic calculation")
		return
	}

	const maxCounter = 4294967296 // mikrotik 32-bit counter bug in wg peers (2^32)

	for _, peer := range peers {
		wgPeer, err := c.mikrotikAdaptor.FetchWgPeer(context.Background(), peer.PeerID)
		if err != nil {
			c.logger.Error("Failed to fetch wireguard peer", zap.String("peerID", peer.PeerID), zap.Error(err))
			continue
		}

		currentTx := utils.ParseStringToInt(wgPeer.TransferTx)
		currentRx := utils.ParseStringToInt(wgPeer.TransferRx)

		deltaTx := calculateDelta(peer.LastTx, currentTx, maxCounter)
		deltaRx := calculateDelta(peer.LastRx, currentRx, maxCounter)

		peer.DownloadUsage += deltaTx
		peer.UploadUsage += deltaRx

		peer.LastTx = currentTx
		peer.LastRx = currentRx

		totalBytes := peer.DownloadUsage + peer.UploadUsage
		if peer.TrafficLimit != nil && totalBytes > *peer.TrafficLimit {
			c.logger.Warn("Peer traffic limit exceeded", zap.String("peerID", peer.PeerID))

			peer.Disabled = true
			_, err := c.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{
				Disabled: strconv.FormatBool(true),
			})
			if err != nil {
				c.logger.Error("Failed to disable peer on Mikrotik", zap.String("peerID", peer.PeerID), zap.Error(err))
			}
		}

		if err := c.db.Save(&peer).Error; err != nil {
			c.logger.Error("Failed to update peer usage in database", zap.String("peerID", peer.PeerID), zap.Error(err))
		}
	}

	c.logger.Info("Peer Traffic calculation job completed")
}

func (c *Calculator) CalculateDailyTraffic() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var interfaces []model.Interface
	if err := c.db.Find(&interfaces).Error; err != nil {
		c.logger.Error("Failed to fetch interfaces from database", zap.Error(err))
		return
	}

	if len(interfaces) == 0 {
		c.logger.Info("No interfaces found, skipping daily traffic calculation")
		return
	}

	for _, iface := range interfaces {
		wgInterface, err := c.mikrotikAdaptor.FetchWgPeer(context.Background(), iface.InterfaceID)
		if err != nil {
			c.logger.Error("Failed to fetch WireGuard interface", zap.String("interfaceID", iface.InterfaceID), zap.Error(err))
			continue
		}

		currentDownload := utils.ParseStringToInt(wgInterface.TransferTx)
		currentUpload := utils.ParseStringToInt(wgInterface.TransferRx)
		currentTotal := currentDownload + currentUpload

		var lastTraffic model.Traffic
		err = c.db.
			Where("interface_id = ?", iface.ID).
			Order("created_at DESC").
			First(&lastTraffic).Error

		var diffDownload, diffUpload, diffTotal int64
		if err == nil {
			diffDownload = currentDownload - lastTraffic.DownloadUsage
			diffUpload = currentUpload - lastTraffic.UploadUsage
			diffTotal = currentTotal - lastTraffic.TotalUsage
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			diffDownload = currentDownload
			diffUpload = currentUpload
			diffTotal = currentTotal
		} else {
			c.logger.Error("Failed to fetch previous traffic record", zap.String("interfaceID", iface.InterfaceID), zap.Error(err))
			continue
		}

		newTraffic := model.Traffic{
			InterfaceID:   iface.ID,
			DownloadUsage: diffDownload,
			UploadUsage:   diffUpload,
			TotalUsage:    diffTotal,
		}

		if err := c.db.Create(&newTraffic).Error; err != nil {
			c.logger.Error("Failed to save daily traffic data", zap.String("interfaceID", iface.InterfaceID), zap.Error(err))
			continue
		}
	}

	c.logger.Info("Daily traffic calculation completed")
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

	currentTx := utils.ParseStringToInt(wgPeer.TransferTx)
	currentRx := utils.ParseStringToInt(wgPeer.TransferRx)

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
