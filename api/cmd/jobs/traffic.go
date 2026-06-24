package traffic

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PeerUsageNotifier interface {
	NotifyPeerUsage(ctx context.Context, peerName, telegramUsername string, percent int64, totalUsage, limit int64) error
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
		wgInterface, err := c.mikrotikAdaptor.FetchInterface(context.Background(), iface.InterfaceID)
		if err != nil {
			c.logger.Error("Failed to fetch WireGuard interface", zap.String("interfaceID", iface.InterfaceID), zap.Error(err))
			continue
		}

		currentDownload := utils.ParseStringToInt(wgInterface.TxByte)
		currentUpload := utils.ParseStringToInt(wgInterface.RxByte)
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

	var peer model.Peer
	if err := c.db.Where("id = ?", id).First(&peer).Error; err != nil {
		c.logger.Error("Failed to find peer in DB", zap.Uint("id", id), zap.Error(err))
		return err
	}

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
		peer.FirstNotify = false
		peer.SecondNotify = false
		peer.ThirdNotify = false

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

	var peers []model.Peer
	if err := c.db.Find(&peers).Error; err != nil {
		c.logger.Error("Failed to find peers in DB", zap.Error(err))
		return err
	}

	wgPeers, err := c.mikrotikAdaptor.FetchWgPeers(context.Background())
	if err != nil {
		c.logger.Error("Failed to fetch peers from Mikrotik", zap.Error(err))
		return err
	}

	wgPeerMap := make(map[string]mikrotik.WireGuardPeer)
	for _, wgPeer := range wgPeers {
		wgPeerMap[wgPeer.ID] = wgPeer
	}

	for _, peer := range peers {
		wgPeer, found := wgPeerMap[peer.PeerID]
		if !found {
			continue
		}

		currentTx := utils.ParseStringToInt(wgPeer.TransferTx)
		currentRx := utils.ParseStringToInt(wgPeer.TransferRx)

		peer.DownloadUsage = 0
		peer.UploadUsage = 0
		peer.LastTx = currentTx
		peer.LastRx = currentRx
		peer.FirstNotify = false
		peer.SecondNotify = false
		peer.ThirdNotify = false

		if err := c.db.Save(&peer).Error; err != nil {
			c.logger.Error("Failed to reset peer usage", zap.String("peerID", peer.PeerID), zap.Error(err))
			return err
		}
	}

	c.logger.Info("Peer usages reset successfully")
	return nil
}

func (c *Calculator) fetchPeers() ([]model.Peer, error) {
	var peers []model.Peer
	if err := c.db.Find(&peers).Error; err != nil {
		c.logger.Error("Failed to fetch peers from database", zap.Error(err))
		return nil, err
	}
	return peers, nil
}

func (c *Calculator) processPeerTraffic(peer model.Peer, maxCounter int64) {
	wgPeer, err := c.mikrotikAdaptor.FetchWgPeer(context.Background(), peer.PeerID)
	if err != nil {
		c.logger.Error("Failed to fetch wireguard peer", zap.String("peerID", peer.PeerID), zap.Error(err))
		return
	}

	currentTx := utils.ParseStringToInt(wgPeer.TransferTx)
	currentRx := utils.ParseStringToInt(wgPeer.TransferRx)

	deltaTx, deltaRx, resetDetected := c.calculatePeerDeltas(peer, currentTx, currentRx, maxCounter)
	if delta := deltaTx + deltaRx; delta > 0 {
		c.accumulateTotalTraffic(delta)
	}
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
	c.persistPeerTraffic(peer, updates)
}

func (c *Calculator) calculatePeerDeltas(peer model.Peer, currentTx, currentRx, maxCounter int64) (int64, int64, bool) {
	deltaTx, resetTx := calculateDelta(peer.LastTx, currentTx, maxCounter)
	deltaRx, resetRx := calculateDelta(peer.LastRx, currentRx, maxCounter)
	return deltaTx, deltaRx, resetTx || resetRx
}

func (c *Calculator) applyPeerTrafficLimit(peer *model.Peer, updates map[string]interface{}) {
	if peer.TrafficLimit != nil && (peer.DownloadUsage+peer.UploadUsage) > *peer.TrafficLimit {
		c.logger.Warn("Peer traffic limit exceeded", zap.String("peerID", peer.PeerID))
		peer.Disabled = true
		updates["disabled"] = true

		_, err := c.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{
			Disabled: strconv.FormatBool(true),
		})
		if err != nil {
			c.logger.Error("Failed to disable peer on Mikrotik", zap.String("peerID", peer.PeerID), zap.Error(err))
		}
	}
}

func (c *Calculator) applyPeerTrafficNotifications(peer *model.Peer, updates map[string]interface{}) {
	if c.notifier == nil || peer.TrafficLimit == nil || peer.TelegramUsername == nil {
		return
	}

	username := strings.TrimSpace(*peer.TelegramUsername)
	if username == "" {
		return
	}

	limit := *peer.TrafficLimit
	if limit <= 0 {
		return
	}

	totalUsage := peer.DownloadUsage + peer.UploadUsage
	percent := (totalUsage * 100) / limit

	c.notifyThreshold(peer, updates, username, percent, totalUsage, limit, 80, "first_notify", &peer.FirstNotify)
	c.notifyThreshold(peer, updates, username, percent, totalUsage, limit, 90, "second_notify", &peer.SecondNotify)
	c.notifyThreshold(peer, updates, username, percent, totalUsage, limit, 100, "third_notify", &peer.ThirdNotify)
}

func (c *Calculator) notifyThreshold(peer *model.Peer, updates map[string]interface{}, username string, percent, totalUsage, limit, threshold int64, updateKey string, notified *bool) {
	if percent < threshold || *notified {
		return
	}

	err := c.notifier.NotifyPeerUsage(context.Background(), peer.Name, username, percent, totalUsage, limit)
	if err != nil {
		c.logger.Error("Failed to send peer usage notification", zap.String("peerID", peer.PeerID), zap.Error(err))
		return
	}

	*notified = true
	updates[updateKey] = true
}

func (c *Calculator) persistPeerTraffic(peer model.Peer, updates map[string]interface{}) {
	if err := c.db.Model(&model.Peer{}).Where("id = ?", peer.ID).Updates(updates).Error; err != nil {
		c.logger.Error("Failed to update peer usage in database", zap.String("peerID", peer.PeerID), zap.Error(err))
	}
}

func (c *Calculator) accumulateTotalTraffic(delta int64) {
	var totalTraffic model.TotalTrafficUsage
	err := c.db.FirstOrCreate(&totalTraffic, model.TotalTrafficUsage{Model: model.Model{ID: model.TotalTrafficUsageSingletonID}}).Error
	if err != nil {
		c.logger.Error("Failed to fetch total traffic usage record", zap.Error(err))
		return
	}

	if err := c.db.Model(&model.TotalTrafficUsage{}).
		Where("id = ?", model.TotalTrafficUsageSingletonID).
		UpdateColumn("total_usage", gorm.Expr("total_usage + ?", delta)).Error; err != nil {
		c.logger.Error("Failed to accumulate total traffic usage", zap.Error(err))
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
	if current >= prev {
		return current - prev, false
	}

	// If counters are already beyond the expected 32-bit max, treat this as a reset.
	if maxCounter <= 0 || prev > maxCounter || current > maxCounter {
		return current, true
	}

	// If the counter moved backwards by a large amount, assume a wrap.
	if (prev - current) > (maxCounter / 2) {
		return (maxCounter - prev) + current, false
	}

	// Otherwise treat it as a reset to avoid a large, incorrect delta.
	return current, true
}
