package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/utils"
	"github.com/maahdima/mwp/api/utils/timehelper"
	"github.com/maahdima/mwp/api/utils/wireguard"
)

type WgPeer struct {
	db              *gorm.DB
	mikrotikAdaptor *mikrotik.Adaptor
	scheduler       *Scheduler
	queue           *Queue
	configGenerator *ConfigGenerator
	qrCodeGenerator *QRCodeGenerator
	logger          *zap.Logger
}

func NewWGPeer(db *gorm.DB, mikrotikAdaptor *mikrotik.Adaptor, scheduler *Scheduler, queue *Queue, configGenerator *ConfigGenerator) *WgPeer {
	return &WgPeer{
		db:              db,
		mikrotikAdaptor: mikrotikAdaptor,
		scheduler:       scheduler,
		queue:           queue,
		configGenerator: configGenerator,
		logger:          zap.L().Named("WgPeerService"),
	}
}

func (w *WgPeer) TogglePeerStatus(id uint) error {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
	}

	disabled := strconv.FormatBool(!peer.Disabled)

	wgPeer := mikrotik.WireGuardPeer{
		Disabled: disabled,
	}
	wgScheduler := mikrotik.Scheduler{
		Disabled: disabled,
	}
	wgQueue := mikrotik.Queue{
		Disabled: disabled,
	}

	if _, err := w.mikrotikAdaptor.UpdateWgPeer(context.Background(), peer.PeerID, wgPeer); err != nil {
		w.logger.Error("failed to update wireguard peer in Mikrotik", zap.Error(err))
		return fmt.Errorf("failed to update wireguard peer: %w", err)
	}

	if peer.SchedulerID != nil {
		if _, err := w.mikrotikAdaptor.UpdateScheduler(context.Background(), *peer.SchedulerID, wgScheduler); err != nil {
			w.logger.Error("failed to update scheduler for wireguard peer", zap.Error(err))
			return fmt.Errorf("failed to update scheduler: %w", err)
		}
	}
	if peer.QueueID != nil {
		if _, err := w.mikrotikAdaptor.UpdateSimpleQueue(context.Background(), *peer.QueueID, wgQueue); err != nil {
			w.logger.Error("failed to update queue for wireguard peer", zap.Error(err))
			return fmt.Errorf("failed to update queue: %w", err)
		}
	}

	if err := w.db.Model(&peer).Update("disabled", disabled).Error; err != nil {
		w.logger.Error("failed to update peer status in database", zap.Error(err))
		return fmt.Errorf("failed to update peer status in database: %w", err)
	}

	return nil
}

func (w *WgPeer) GetPeerKeys() (*schema.PeerKeyResponse, error) {
	privateKey, publicKey, err := w.GenerateKeys()
	if err != nil {
		w.logger.Error("failed to generate peer keys", zap.Error(err))
		return nil, err
	}

	return &schema.PeerKeyResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func (w *WgPeer) GetPeerShareStatus(id uint) (*schema.PeerShareStatusResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Error("peer not found in database", zap.Uint("id", id))
			return nil, fmt.Errorf("peer not found: %w", err)
		}
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return nil, err
	}

	if !peer.IsShared {
		return &schema.PeerShareStatusResponse{
			IsShared:   false,
			ShareLink:  nil,
			ExpireTime: nil,
		}, nil
	}

	appCfg := config.GetAppConfig()

	// TODO: https
	shareLink := fmt.Sprintf("http://%s:%s/share?shareId=%s", appCfg.Host, appCfg.Port, peer.UUID)

	return &schema.PeerShareStatusResponse{
		IsShared:   peer.IsShared,
		ShareLink:  &shareLink,
		ExpireTime: peer.ShareExpireTime,
	}, nil
}

func (w *WgPeer) TogglePeerShareStatus(id uint) error {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
	}

	isShared := !peer.IsShared

	if err := w.db.Model(&peer).Update("is_shared", isShared).Error; err != nil {
		w.logger.Error("failed to update peer share status in database", zap.Error(err))
		return fmt.Errorf("failed to update peer share status: %w", err)
	}

	return nil
}

func (w *WgPeer) UpdatePeerShareExpireTime(id uint, expireTime *string) error {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
	}

	if !peer.IsShared {
		w.logger.Error("peer is not shared, cannot set expire time", zap.Uint("id", id))
		return fmt.Errorf("peer is not shared, cannot set expire time")
	}

	if err := w.db.Model(&peer).Update("share_expire_time", expireTime).Error; err != nil {
		w.logger.Error("failed to update peer share expire time in database", zap.Error(err))
		return fmt.Errorf("failed to update peer share expire time: %w", err)
	}

	return nil
}

func (w *WgPeer) GetPeerDetails(uuid string) (*schema.PeerDetailsResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "uuid = ?", uuid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.logger.Error("peer not found in database", zap.String("uuid", uuid))
			return nil, err
		}
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return nil, err
	}

	isSharable := utils.IsPeerSharable(peer.IsShared, peer.ShareExpireTime)
	if !isSharable {
		return nil, common.ErrPeerNotShared
	}

	totalUsage := peer.DownloadUsage + peer.UploadUsage

	var usagePercent, trafficLimit *string
	if peer.TrafficLimit != nil {
		trafficLimit = utils.Ptr(utils.BytesToGB(*peer.TrafficLimit))
		percent := float64(totalUsage) / float64(*peer.TrafficLimit) * 100
		usagePercent = utils.Ptr(fmt.Sprintf("%.1f", percent))
	}

	return &schema.PeerDetailsResponse{
		Name:          peer.Name,
		TrafficLimit:  trafficLimit,
		ExpireTime:    peer.ExpireTime,
		DownloadUsage: utils.BytesToGB(peer.DownloadUsage),
		UploadUsage:   utils.BytesToGB(peer.UploadUsage),
		TotalUsage:    utils.BytesToGB(totalUsage),
		UsagePercent:  usagePercent,
	}, nil
}

func (w *WgPeer) GetPeers() (*[]schema.PeerResponse, error) {
	var dbPeers []model.Peer
	if err := w.db.Order("created_at ASC").Find(&dbPeers).Error; err != nil {
		w.logger.Error("failed to get peers from database", zap.Error(err))
		return nil, err
	}

	var wgPeers []schema.PeerResponse
	for _, dbPeer := range dbPeers {
		peer, err := w.mikrotikAdaptor.FetchWgPeer(context.Background(), dbPeer.PeerID)
		if err != nil {
			w.logger.Error("failed to fetch wireguard peer from Mikrotik", zap.String("peerID", dbPeer.PeerID), zap.Error(err))
			continue
		}

		wgPeer := w.transformPeerToResponse(dbPeer)

		var duration time.Duration
		if peer.LastHandshake != nil {
			duration, err = time.ParseDuration(*peer.LastHandshake)
			if err != nil {
				w.logger.Error("failed to parse last handshake duration", zap.Error(err))
				return nil, fmt.Errorf("failed to parse last handshake duration: %w", err)
			}

			if duration < 150*time.Second {
				wgPeer.IsOnline = true
			}
		}

		wgPeers = append(wgPeers, wgPeer)
	}

	return &wgPeers, nil
}

func (w *WgPeer) CreatePeer(req *schema.CreatePeerRequest) (*schema.PeerResponse, error) {
	wgInterface, err := w.mikrotikAdaptor.FetchWgInterface(context.Background(), req.InterfaceId)
	if err != nil {
		return nil, err
	}

	var existingPeer model.Peer
	if err := w.db.Where("allowed_address = ?", req.AllowedAddress).First(&existingPeer).Error; err == nil {
		return nil, fmt.Errorf("allowed address %s is already in use by peer %s", req.AllowedAddress, existingPeer.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		w.logger.Error("failed to check allowed address uniqueness", zap.Error(err))
		return nil, err
	}

	wgPeer := &mikrotik.WireGuardPeer{
		Comment:        req.Comment,
		Name:           req.Name,
		AllowedAddress: req.AllowedAddress,
		Interface:      req.Interface,
		PresharedKey:   req.PresharedKey,
		PrivateKey:     &req.PrivateKey,
		PublicKey:      req.PublicKey,
	}

	mtPeer, err := w.mikrotikAdaptor.CreateWgPeer(context.Background(), *wgPeer)
	if err != nil {
		return nil, err
	}

	schedulerId, err := w.scheduler.createScheduler(mtPeer.ID, mtPeer.Name, req.ExpireTime)
	if err != nil {
		return nil, err
	}

	queueId, err := w.queue.createQueue(mtPeer.Name, mtPeer.AllowedAddress, req.DownloadBandwidth, req.UploadBandwidth)
	if err != nil {
		return nil, err
	}

	var timeString = common.DefaultKeepalive
	if req.PersistentKeepAlive != nil {
		parsedTime, err := timehelper.ParseTime(*req.PersistentKeepAlive)
		if err != nil {
			w.logger.Error("failed to parse persistent keepalive time", zap.Error(err))
			return nil, err
		}
		timeString = strconv.Itoa(parsedTime)
	}

	disabled, err := strconv.ParseBool(mtPeer.Disabled)
	if err != nil {
		w.logger.Error("failed to parse disabled field from Mikrotik peer", zap.Error(err))
		return nil, err
	}

	trafficLimit := utils.GBToBytes(utils.DerefString(req.TrafficLimit))

	dbPeer := model.Peer{
		UUID:                uuid.New().String(),
		PeerID:              mtPeer.ID,
		Disabled:            disabled,
		Comment:             mtPeer.Comment,
		Name:                mtPeer.Name,
		PrivateKey:          *mtPeer.PrivateKey,
		PublicKey:           mtPeer.PublicKey,
		Interface:           mtPeer.Interface,
		AllowedAddress:      mtPeer.AllowedAddress,
		Endpoint:            req.Endpoint,
		EndpointPort:        wgInterface.ListenPort,
		PersistentKeepalive: timeString,
		SchedulerID:         schedulerId,
		QueueID:             queueId,
		TrafficLimit:        &trafficLimit,
		DownloadBandwidth:   req.DownloadBandwidth,
		UploadBandwidth:     req.UploadBandwidth,
	}
	if err := w.db.Create(&dbPeer).Error; err != nil {
		w.logger.Error("failed to create peer in database", zap.Error(err))
		return nil, err
	}

	configData := fmt.Sprintf(wireguard.Template, req.PrivateKey, dbPeer.AllowedAddress, common.DefaultDns, wgInterface.PublicKey, dbPeer.Endpoint, dbPeer.EndpointPort, common.AllowedIpsIncludeLocal, dbPeer.PersistentKeepalive)

	err = w.configGenerator.BuildPeerConfig(
		configData,
		dbPeer.UUID,
	)
	if err != nil {
		return nil, err
	}

	err = w.qrCodeGenerator.BuildPeerQRCode(configData, dbPeer.UUID)
	if err != nil {
		return nil, err
	}

	transformedPeer := w.transformPeerToResponse(dbPeer)
	return &transformedPeer, nil
}

func (w *WgPeer) UpdatePeer(id uint, req *schema.UpdatePeerRequest) (*schema.PeerResponse, error) {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to get peer from database", zap.Error(err))
		return nil, err
	}

	if err := w.updateMikrotikPeer(peer.PeerID, req); err != nil {
		return nil, err
	}

	schedulerID, err := w.handleScheduler(&peer, req)
	if err != nil {
		return nil, err
	}

	queueID, err := w.handleQueue(&peer, req)
	if err != nil {
		return nil, err
	}

	updateData := w.preparePeerUpdate(req, schedulerID, queueID)
	if err := w.db.Model(&peer).Updates(updateData).Error; err != nil {
		return nil, err
	}

	transformed := w.transformPeerToResponse(peer)
	return &transformed, nil
}

func (w *WgPeer) DeletePeer(id uint) error {
	var peer model.Peer
	if err := w.db.First(&peer, "id = ?", id).Error; err != nil {
		w.logger.Error("failed to find peer in database", zap.Error(err))
		return fmt.Errorf("peer not found: %w", err)
	}

	if err := w.scheduler.deleteScheduler(peer.SchedulerID); err != nil {
		return fmt.Errorf("failed to delete scheduler: %w", err)
	}

	if err := w.queue.deleteQueue(peer.QueueID); err != nil {
		return fmt.Errorf("failed to delete simple queue: %w", err)
	}

	if err := w.mikrotikAdaptor.DeleteWgPeer(context.Background(), peer.PeerID); err != nil {
		w.logger.Error("failed to delete wireguard peer from Mikrotik", zap.Error(err))
		return fmt.Errorf("failed to delete wireguard peer: %w", err)
	}

	err := w.qrCodeGenerator.RemovePeerQRCode(id)
	if err != nil {
		w.logger.Error("failed to remove QR Code file", zap.Error(err))
		return err
	}

	err = w.configGenerator.RemovePeerConfig(id)
	if err != nil {
		w.logger.Error("failed to remove peer config", zap.Error(err))
		return err
	}

	if err := w.db.Unscoped().Delete(&peer).Error; err != nil {
		w.logger.Error("failed to delete peer from database", zap.Error(err))
		return fmt.Errorf("failed to delete peer from database: %w", err)
	}

	return nil
}

func (w *WgPeer) GetPeersData() (*schema.PeerStatsResponse, error) {
	peers, err := w.mikrotikAdaptor.FetchWgPeers(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wireguard peers from Mikrotik: %w", err)
	}

	type peerWithDuration struct {
		peer     mikrotik.WireGuardPeer
		duration time.Duration
	}

	var allOnlinePeers []peerWithDuration
	var disabledPeers []mikrotik.WireGuardPeer

	for _, peer := range peers {
		if peer.LastHandshake != nil {
			duration, err := time.ParseDuration(*peer.LastHandshake)
			if err != nil {
				w.logger.Error("failed to parse last handshake duration", zap.Error(err))
				return nil, fmt.Errorf("failed to parse last handshake duration: %w", err)
			}
			allOnlinePeers = append(allOnlinePeers, peerWithDuration{
				peer:     peer,
				duration: duration,
			})
		}
		if peer.Disabled == "true" {
			disabledPeers = append(disabledPeers, peer)
		}
	}

	onlinePeersCount := len(allOnlinePeers)

	sort.Slice(allOnlinePeers, func(i, j int) bool {
		return allOnlinePeers[i].duration < allOnlinePeers[j].duration
	})

	var wgPeers []schema.RecentOnlinePeers
	for i, item := range allOnlinePeers {
		if i >= 5 {
			break
		}
		wgPeers = append(wgPeers, schema.RecentOnlinePeers{
			Name:     item.peer.Name,
			LastSeen: time.Unix(int64(item.duration.Seconds()), 0).UTC().Format("15:04:05"),
		})
	}

	return &schema.PeerStatsResponse{
		RecentOnlinePeers: &wgPeers,
		TotalPeers:        len(peers),
		OnlinePeers:       onlinePeersCount,
		OfflinePeers:      len(peers) - onlinePeersCount,
		DisabledPeers:     len(disabledPeers),
	}, nil
}

func (w *WgPeer) GenerateKeys() (privateKey, publicKey string, err error) {
	privKey, privateKey, err := wireguard.GeneratePrivateKey()
	if err != nil {
		w.logger.Error("failed to generate private key", zap.Error(err))
		return
	}

	publicKey, err = wireguard.GeneratePublicKey(privKey)
	if err != nil {
		w.logger.Error("failed to generate public key from private key", zap.Error(err))
		return "", "", err
	}

	return
}

func (w *WgPeer) updateMikrotikPeer(peerID string, req *schema.UpdatePeerRequest) error {
	wgPeer := mikrotik.WireGuardPeer{}

	if req.Disabled != nil {
		disabledStr := strconv.FormatBool(*req.Disabled)
		wgPeer.Disabled = disabledStr
	}
	if req.Comment != nil {
		wgPeer.Comment = req.Comment
	}

	wgPeer.Name = req.Name
	wgPeer.AllowedAddress = req.AllowedAddress

	if req.PersistentKeepAlive != nil {
		wgPeer.PersistentKeepAlive = req.PersistentKeepAlive
	}
	if req.PresharedKey != nil {
		wgPeer.PresharedKey = req.PresharedKey
	}

	_, err := w.mikrotikAdaptor.UpdateWgPeer(context.Background(), peerID, wgPeer)
	if err != nil {
		w.logger.Error("failed to update wireguard peer in Mikrotik", zap.Error(err))
	}

	return err
}

func (w *WgPeer) handleScheduler(peer *model.Peer, req *schema.UpdatePeerRequest) (*string, error) {
	if req.ExpireTime == nil && peer.SchedulerID != nil {
		err := w.scheduler.deleteScheduler(peer.SchedulerID)
		if err != nil {
			w.logger.Error("failed to delete scheduler for wireguard peer", zap.Error(err))
			return peer.SchedulerID, err
		}
		return nil, nil
	}

	if req.ExpireTime != nil && peer.SchedulerID == nil {
		return w.scheduler.createScheduler(peer.PeerID, peer.Name, req.ExpireTime)
	}

	if req.ExpireTime != nil {
		err := w.scheduler.updateScheduler(peer.SchedulerID, req.ExpireTime)
		if err != nil {
			w.logger.Error("failed to update scheduler for WireGuard peer", zap.Error(err))
			return peer.SchedulerID, err
		}
	}

	return peer.SchedulerID, nil
}

func (w *WgPeer) handleQueue(peer *model.Peer, req *schema.UpdatePeerRequest) (*string, error) {
	download := req.DownloadBandwidth
	upload := req.UploadBandwidth
	queueID := peer.QueueID

	if download == nil && upload == nil {
		if queueID != nil {
			err := w.queue.deleteQueue(queueID)
			if err != nil {
				w.logger.Error("failed to delete queue for wireguard peer", zap.Error(err))
				return queueID, err
			}
		}
		return nil, nil
	}

	if queueID == nil {
		newQueueID, err := w.queue.createQueue(peer.Name, peer.AllowedAddress, download, upload)
		if err != nil {
			w.logger.Error("failed to create queue for wireguard peer", zap.Error(err))
			return nil, err
		}
		return newQueueID, nil
	}

	err := w.queue.updateQueue(queueID, download, upload)
	if err != nil {
		w.logger.Error("failed to update queue for wireguard peer", zap.Error(err))
		return queueID, err
	}

	return queueID, nil
}

func (w *WgPeer) preparePeerUpdate(req *schema.UpdatePeerRequest, schedulerID, queueID *string) map[string]interface{} {
	updateData := map[string]interface{}{}

	trafficLimit := utils.GBToBytes(utils.DerefString(req.TrafficLimit))
	if trafficLimit > 0 {
		updateData["traffic_limit"] = trafficLimit
	} else {
		updateData["traffic_limit"] = nil
	}

	updateData["disabled"] = req.Disabled
	updateData["comment"] = req.Comment
	updateData["name"] = req.Name
	updateData["allowed_address"] = req.AllowedAddress
	updateData["persistent_keepalive"] = req.PersistentKeepAlive
	updateData["expire_time"] = req.ExpireTime
	updateData["download_bandwidth"] = req.DownloadBandwidth
	updateData["upload_bandwidth"] = req.UploadBandwidth
	updateData["scheduler_id"] = schedulerID
	updateData["queue_id"] = queueID

	return updateData
}

func (w *WgPeer) transformPeerToResponse(peer model.Peer) schema.PeerResponse {
	statuses := w.transformPeerStatus(peer)

	var trafficLimit *string

	if peer.TrafficLimit == nil {
		trafficLimit = nil
	} else {
		trafficLimit = utils.Ptr(utils.BytesToGB(*peer.TrafficLimit))
	}

	return schema.PeerResponse{
		Id:                peer.ID,
		UUID:              peer.UUID,
		Disabled:          peer.Disabled,
		Comment:           peer.Comment,
		Name:              peer.Name,
		Interface:         peer.Interface,
		AllowedAddress:    peer.AllowedAddress,
		TrafficLimit:      trafficLimit,
		ExpireTime:        peer.ExpireTime,
		DownloadBandwidth: peer.DownloadBandwidth,
		UploadBandwidth:   peer.UploadBandwidth,
		TotalUsage:        utils.BytesToGB(peer.DownloadUsage + peer.UploadUsage),
		Status:            statuses,
		IsShared:          peer.IsShared,
	}
}

func (w *WgPeer) transformPeerStatus(peer model.Peer) []schema.PeerStatus {
	var peerStatus []schema.PeerStatus

	if peer.Disabled {
		peerStatus = append(peerStatus, schema.InactivePeer)
	} else {
		peerStatus = append(peerStatus, schema.ActivePeer)
	}

	if peer.ExpireTime != nil {
		expireTime, err := time.Parse("2006-01-02", *peer.ExpireTime)
		if err == nil && time.Now().After(expireTime) {
			peerStatus = append(peerStatus, schema.ExpiredPeer)
		}
	}

	if peer.TrafficLimit != nil {
		totalUsedTraffic := peer.DownloadUsage + peer.UploadUsage
		if totalUsedTraffic > *peer.TrafficLimit {
			peerStatus = append(peerStatus, schema.SuspendedPeer)
		}
	}

	return peerStatus
}
