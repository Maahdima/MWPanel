package traffic

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PeerUsageNotifier interface {
	Enabled() bool
	NotifyPeerUsage(ctx context.Context, peer *model.Peer, percent, totalUsage, limit int64) error
	NotifyPeerExpiry(ctx context.Context, peer *model.Peer, daysLeft int) error
}

type Calculator struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	mu              *sync.Mutex // avoid race condition between traffic job and reset
	logger          *zap.Logger
	notifier        PeerUsageNotifier
}

func NewTrafficCalculator(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, notifier PeerUsageNotifier) *Calculator {
	return &Calculator{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		mu:              &sync.Mutex{},
		logger:          zap.L().Named("TrafficCalculatorJob"),
		notifier:        notifier,
	}
}

func (c *Calculator) CalculatePeerTraffic() {
	c.mu.Lock()

	defer c.mu.Unlock()

	peers, err := c.fetchPeers()

	if err != nil {
		return
	}

	if len(peers) == 0 {
		c.logger.Info("No peers found, skipping traffic calculation")

		return
	}

	const maxCounter = 4294967296 // mikrotik 32-bit counter bug in wg peers (2^32)

	for _, peer := range peers {
		c.processPeerTraffic(peer, maxCounter)
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
		adaptor, err := c.mikrotikAdaptor.ForServer(iface.ServerID)

		if err != nil {
			continue
		}

		wgInterface, err := adaptor.FetchInterface(context.Background(), iface.InterfaceID)

		if err != nil {
			c.logger.Error("Failed to fetch WireGuard interface", zap.String("interfaceID", iface.InterfaceID), zap.Error(err))

			continue
		}

		// MikroTik tx-byte is traffic sent by the interface (download for peers),
		// rx-byte is traffic received (upload from peers).
		currentDownload := utils.ParseStringToInt(wgInterface.TxByte)

		currentUpload := utils.ParseStringToInt(wgInterface.RxByte)

		var diffDownload, diffUpload int64

		if iface.DailyTrafficReady {
			diffDownload = interfaceDailyDelta(iface.LastTx, currentDownload)

			diffUpload = interfaceDailyDelta(iface.LastRx, currentUpload)
		}

		if err := c.db.Model(&model.Interface{}).
			Where("id = ?", iface.ID).
			Updates(map[string]interface{}{
				"last_tx":             currentDownload,
				"last_rx":             currentUpload,
				"daily_traffic_ready": true,
			}).Error; err != nil {
			c.logger.Error("Failed to update interface traffic counters", zap.String("interfaceID", iface.InterfaceID), zap.Error(err))

			continue
		}

		// First sample after deploy/reset only establishes the baseline.
		if !iface.DailyTrafficReady {
			c.logger.Info("Established daily traffic baseline",
				zap.String("interfaceID", iface.InterfaceID),
				zap.Int64("lastTx", currentDownload),
				zap.Int64("lastRx", currentUpload),
			)

			continue
		}

		newTraffic := model.Traffic{
			InterfaceID:   iface.ID,
			DownloadUsage: diffDownload,
			UploadUsage:   diffUpload,
			TotalUsage:    diffDownload + diffUpload,
		}

		if err := c.db.Create(&newTraffic).Error; err != nil {
			c.logger.Error("Failed to save daily traffic data", zap.String("interfaceID", iface.InterfaceID), zap.Error(err))

			continue
		}
	}

	c.logger.Info("Daily traffic calculation completed")
}

// interfaceDailyDelta returns bytes transferred since the previous absolute counter.
// A drop in the counter is treated as a reset (reboot / interface recreate).
func interfaceDailyDelta(prev, current int64) int64 {
	if current >= prev {
		return current - prev
	}

	return current
}

func (c *Calculator) ResetPeerUsage(id uint) error {
	c.mu.Lock()

	defer c.mu.Unlock()

	var peer model.Peer

	if err := c.db.Where("id = ?", id).First(&peer).Error; err != nil {
		c.logger.Error("Failed to find peer in DB", zap.Uint("id", id), zap.Error(err))

		return err
	}

	if peer.ResellerID != nil || peer.PurchaseID != nil {
		return fmt.Errorf("owned peer requires a billed reset")
	}

	selected, err := c.forServer(peer.ServerID)

	if err != nil {
		return err
	}

	c = selected

	wgPeer, err := c.mikrotikAdaptor.FetchWgPeer(context.Background(), peer.PeerID)

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

		peer.TrafficFirstNotify = false

		peer.TrafficSecondNotify = false

		peer.TrafficThirdNotify = false

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

func (c *Calculator) ResetPeerUsages() error {
	c.mu.Lock()

	defer c.mu.Unlock()

	var owned int64

	if err := c.db.Model(&model.Peer{}).Where("reseller_id IS NOT NULL OR purchase_id IS NOT NULL").Count(&owned).Error; err != nil {
		return err
	}

	if owned > 0 {
		return fmt.Errorf("owned peers require individual billed resets")
	}

	peers, err := c.fetchPeers()

	if err != nil {
		return err
	}

	for _, peer := range peers {
		if peer.ResellerID != nil || peer.PurchaseID != nil {
			return fmt.Errorf("owned peers require individual billed resets")
		}
	}

	observed := c.observedPeers(peers)

	// Preflight every router/peer before resetting any database row.
	for _, peer := range peers {
		if _, ok := observed[peer.ServerID][peer.PeerID]; !ok {
			return fmt.Errorf("peer %d is unavailable on server %d", peer.ID, peer.ServerID)
		}
	}

	return c.db.Transaction(func(tx *gorm.DB) error {
		for _, peer := range peers {
			snapshot := observed[peer.ServerID][peer.PeerID]

			updates := map[string]interface{}{"download_usage": int64(0), "upload_usage": int64(0), "last_tx": utils.ParseStringToInt(snapshot.TransferTx), "last_rx": utils.ParseStringToInt(snapshot.TransferRx), "traffic_first_notify": false, "traffic_second_notify": false, "traffic_third_notify": false}

			if err := tx.Model(&model.Peer{}).Where("id = ?", peer.ID).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (c *Calculator) fetchPeers() ([]model.Peer, error) {
	var peers []model.Peer

	if err := c.db.Where("purchase_id IS NULL").Find(&peers).Error; err != nil {
		c.logger.Error("Failed to fetch peers from database", zap.Error(err))

		return nil, err
	}

	return peers, nil
}

func (c *Calculator) processPeerTraffic(peer model.Peer, maxCounter int64) {
	now := time.Now().Unix()

	lease := c.db.Model(&model.Peer{}).Where("id = ? AND purchase_id IS NULL AND sync_lease_until <= ?", peer.ID, now).UpdateColumn("sync_lease_until", now+120)

	if lease.Error != nil || lease.RowsAffected != 1 {
		return
	}

	defer c.db.Model(&model.Peer{}).Where("id = ? AND sync_lease_until = ?", peer.ID, now+120).UpdateColumn("sync_lease_until", 0)

	if c.db.First(&peer, peer.ID).Error != nil {
		return
	}

	selected, err := c.forServer(peer.ServerID)

	if err != nil {
		return
	}

	c = selected

	wgPeer, err := c.mikrotikAdaptor.FetchWgPeer(context.Background(), peer.PeerID)

	if err != nil {
		c.logger.Error("Failed to fetch wireguard peer", zap.String("peerID", peer.PeerID), zap.Error(err))

		return
	}

	currentTx := utils.ParseStringToInt(wgPeer.TransferTx)

	currentRx := utils.ParseStringToInt(wgPeer.TransferRx)

	deltaTx, deltaRx, resetDetected := c.calculatePeerDeltas(peer, currentTx, currentRx, maxCounter)

	if resetDetected {
		c.logger.Debug("Detected peer counter reset",
			zap.String("peerID", peer.PeerID),
			zap.Int64("prevTx", peer.LastTx),
			zap.Int64("currentTx", currentTx),
			zap.Int64("prevRx", peer.LastRx),
			zap.Int64("currentRx", currentRx),
		)
	}

	peer.DownloadUsage += deltaTx

	peer.UploadUsage += deltaRx

	peer.LastTx = currentTx

	peer.LastRx = currentRx

	updates := map[string]interface{}{
		"download_usage": peer.DownloadUsage,
		"upload_usage":   peer.UploadUsage,
		"last_tx":        peer.LastTx,
		"last_rx":        peer.LastRx,
	}

	c.applyPeerTrafficNotifications(&peer, updates)

	c.applyPeerTrafficLimit(&peer, updates)

	c.persistPeerTraffic(peer, updates, deltaTx+deltaRx)
}

func (c *Calculator) calculatePeerDeltas(peer model.Peer, currentTx, currentRx, maxCounter int64) (int64, int64, bool) {
	deltaTx, resetTx := calculateDelta(peer.LastTx, currentTx, maxCounter)

	deltaRx, resetRx := calculateDelta(peer.LastRx, currentRx, maxCounter)

	return deltaTx, deltaRx, resetTx || resetRx
}

func (c *Calculator) applyPeerTrafficLimit(peer *model.Peer, updates map[string]interface{}) {
	if peer.TrafficLimit != nil && (peer.DownloadUsage+peer.UploadUsage) >= *peer.TrafficLimit {
		c.logger.Warn("Peer traffic limit exceeded", zap.String("peerID", peer.PeerID))

		peer.Disabled = true

		updates["disabled"] = true

		updates["quota_blocked"] = true

		updates["restriction_revision"] = gorm.Expr("restriction_revision + 1")

		updates["sync_next_attempt_at"] = 0

		_, err := c.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{
			Disabled: strconv.FormatBool(true),
		})

		if err != nil {
			c.logger.Error("Failed to disable peer on Mikrotik", zap.String("peerID", peer.PeerID), zap.Error(err))
		}
	}
}

func (c *Calculator) applyPeerTrafficNotifications(peer *model.Peer, updates map[string]interface{}) {
	if peer.Disabled || c.notifier == nil || !c.notifier.Enabled() || peer.TrafficLimit == nil {
		return
	}

	limit := *peer.TrafficLimit

	if limit <= 0 {
		return
	}

	totalUsage := peer.DownloadUsage + peer.UploadUsage

	percent := (totalUsage * 100) / limit

	c.notifyThreshold(peer, updates, percent, totalUsage, limit, 80, "traffic_first_notify", &peer.TrafficFirstNotify)

	c.notifyThreshold(peer, updates, percent, totalUsage, limit, 90, "traffic_second_notify", &peer.TrafficSecondNotify)

	c.notifyThreshold(peer, updates, percent, totalUsage, limit, 100, "traffic_third_notify", &peer.TrafficThirdNotify)
}

func (c *Calculator) notifyThreshold(peer *model.Peer, updates map[string]interface{}, percent, totalUsage, limit, threshold int64, updateKey string, notified *bool) {
	if peer.Disabled || percent < threshold || *notified {
		return
	}

	err := c.notifier.NotifyPeerUsage(context.Background(), peer, percent, totalUsage, limit)

	if err != nil {
		c.logger.Warn("Failed to send peer usage notification", zap.String("peerID", peer.PeerID), zap.Error(err))

		return
	}

	*notified = true

	updates[updateKey] = true
}

// Commit the sample and global counter together, only while this worker still
// owns the peer lease and the router identity has not changed underneath it.
func (c *Calculator) persistPeerTraffic(peer model.Peer, updates map[string]interface{}, delta int64) {
	err := c.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Peer{}).Where("id = ? AND server_id = ? AND peer_id = ? AND sync_lease_until = ?", peer.ID, peer.ServerID, peer.PeerID, peer.SyncLeaseUntil).Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return fmt.Errorf("peer changed during traffic sample")
		}

		if delta > 0 {
			var total model.TotalTrafficUsage

			if err := tx.FirstOrCreate(&total, model.TotalTrafficUsage{Model: model.Model{ID: model.TotalTrafficUsageSingletonID}}).Error; err != nil {
				return err
			}

			return tx.Model(&model.TotalTrafficUsage{}).Where("id = ?", model.TotalTrafficUsageSingletonID).UpdateColumn("total_usage", gorm.Expr("total_usage + ?", delta)).Error
		}

		return nil
	})

	if err != nil {
		c.logger.Error("Failed to commit peer traffic sample", zap.Uint("peer_id", peer.ID), zap.Error(err))
	}
}

func (c *Calculator) ResetTotalTrafficUsage() error {
	c.mu.Lock()

	defer c.mu.Unlock()

	var totalTraffic model.TotalTrafficUsage

	if err := c.db.FirstOrCreate(&totalTraffic, model.TotalTrafficUsage{Model: model.Model{ID: model.TotalTrafficUsageSingletonID}}).Error; err != nil {
		c.logger.Error("Failed to fetch total traffic usage record", zap.Error(err))

		return err
	}

	if err := c.db.Model(&model.TotalTrafficUsage{}).
		Where("id = ?", model.TotalTrafficUsageSingletonID).
		Update("total_usage", 0).Error; err != nil {
		c.logger.Error("Failed to reset total traffic usage", zap.Error(err))

		return err
	}

	c.logger.Info("Total traffic usage reset successfully")

	return nil
}

func calculateDelta(prev, current, maxCounter int64) (int64, bool) {
	return utils.PeerCounterDelta(prev, current, maxCounter)
}
